// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.
package operator_test

import (
	"context"
	_ "embed"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	helmclient "github.com/mittwald/go-helm-client"

	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	pulsarv1alpha1 "github.com/streamnative/resources-operator/api/v1alpha1"

	helmRepo "helm.sh/helm/v3/pkg/repo"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	kubeConfig        string
	snChart           string
	randomSuffix      uint32
	namespaceName     string
	snHelmReleaseName string
	k8sClient         client.Client
	helmClient        helmclient.Client
	snCharSpec        *helmclient.ChartSpec
	repoCacheDir      string
	nodeIP            string

	//go:embed values.yaml
	yamlStr string
)

func TestOperator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Operator Suite")
}

var _ = BeforeSuite(func() {

	r := rand.New(rand.NewSource(time.Now().Unix()))
	randomSuffix = r.Uint32()
	namespaceName = fmt.Sprintf("pulsar-%d", randomSuffix)
	snHelmReleaseName = fmt.Sprintf("sn-platform-helm-%d", randomSuffix)
	snChart = "streamnative/sn-platform"

	snCharSpec = &helmclient.ChartSpec{
		ReleaseName: snHelmReleaseName,
		ChartName:   snChart,
		Namespace:   namespaceName,
		ValuesYaml:  fmt.Sprintf(yamlStr, namespaceName),
	}

	Expect(pulsarv1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())

	k8sClient = makeK8SClusterClient()

	helmClient = makeHelmClient()

	ctx := context.Background()

	nodeIP = getNodeIP(ctx)

	createNamespace(ctx)

	setupHelmRepo()

	helmInstall(ctx, "sn-platform", snHelmReleaseName, snCharSpec)

})

var _ = AfterSuite(func() {
	err := helmClient.UninstallRelease(snCharSpec)
	if err != nil {
		logrus.Info(err)
		Fail(err.Error())
	}
	logrus.Info("pulsar chart uninstalled")

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	err = k8sClient.Delete(context.Background(), namespace)
	if err != nil {
		logrus.Info(err)
		Fail(err.Error())
	}
	logrus.Info("namespace deleted")
	os.RemoveAll(repoCacheDir)
	logrus.Info("temp repo cache dir removed")
})

func makeHelmClient() helmclient.Client {
	logrus.Info("Creating helm client")
	config, err := getClusterConfig()
	if err != nil {
		logrus.Info(err)
		Fail(err.Error())
	}
	repoCacheDir, err = ioutil.TempDir("/tmp/.helmcache/", "repoCache")
	if err != nil {
		logrus.Info(err)
		Fail(err.Error())
	}
	helmClientConfig := &helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        namespaceName,
			RepositoryCache:  repoCacheDir,
			RepositoryConfig: "/tmp/.helmrepo",
		},
		RestConfig: config,
	}
	helmClient, err = helmclient.NewClientFromRestConf(helmClientConfig)
	if err != nil {
		logrus.Info(err)
		Fail(err.Error())
	}
	logrus.Info("Helm client created successfully")
	return helmClient
}

func makeK8SClusterClient() client.Client {
	logrus.Info("Creating k8s client")
	k8sConfig, err := getClusterConfig()
	if err != nil {
		logrus.Info(err)
		Fail(err.Error())
	}

	k8sClient, err = client.New(k8sConfig, client.Options{})
	Expect(err).Should(Succeed())

	logrus.Info("K8s client created successfully")
	return k8sClient
}

func getClusterConfig() (*rest.Config, error) {
	if len(kubeConfig) == 0 {
		kubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", kubeConfig)
}

func createNamespace(ctx context.Context) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	err := k8sClient.Create(ctx, namespace)
	if err != nil && !errors.IsAlreadyExists(err) {
		logrus.Info(err)
		Fail(err.Error())
	}
	logrus.Infof("namespace %s created", namespaceName)
}

func setupHelmRepo() {
	vaultRepoEntry := helmRepo.Entry{
		Name: "banzaicloud-stable",
		URL:  "https://kubernetes-charts.banzaicloud.com",
	}
	snRepoEntry := helmRepo.Entry{
		Name: "streamnative",
		URL:  "https://charts.streamnative.io",
	}

	err := helmClient.AddOrUpdateChartRepo(vaultRepoEntry)
	if err != nil {
		logrus.Info(err)
		Fail(err.Error())
	}

	err = helmClient.AddOrUpdateChartRepo(snRepoEntry)
	if err != nil {
		logrus.Info(err)
		Fail(err.Error())
	}
	logrus.Info("sn chart repo added")

	err = helmClient.UpdateChartRepos()
	if err != nil {
		logrus.Info(err)
		Fail(err.Error())
	}
	logrus.Info("updated sn chart repo")

}

func helmInstall(ctx context.Context, component, releaseName string, chartSpec *helmclient.ChartSpec) {
	logrus.Info("install ", component)
	release, err := helmClient.InstallOrUpgradeChart(ctx, chartSpec)
	if err != nil && !strings.Contains(err.Error(), "Operation cannot be fulfilled on resourcequotas \"gke-resource-quotas\": the object has been modified") {
		logrus.Info(err)
		Fail(err.Error())
	}
	logrus.Infof("%s installed, release name: %s", component, releaseName)
	logrus.Info(release)
}

func getNodeIP(ctx context.Context) string {
	var ip string

	node := &corev1.NodeList{}
	Expect(k8sClient.List(ctx, node)).Should(Succeed())
	for _, i := range node.Items[0].Status.Addresses {
		if i.Type == corev1.NodeExternalIP {
			logrus.Info("node external ip is ", i.Address)
			ip = i.Address
			break
		}
		if i.Type == corev1.NodeInternalIP {
			logrus.Info("node internal ip is ", i.Address)
			ip = i.Address
		}

	}

	return ip
}
