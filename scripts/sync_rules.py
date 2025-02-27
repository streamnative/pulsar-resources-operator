#!/usr/bin/env python3
import re
import sys

def sync_rules(source_file, target_file):
    """
    Sync the rules section from source_file to target_file
    
    Args:
        source_file: The path to the source file
        target_file: The path to the target file
    """
    # Read the source file and extract the rules section
    with open(source_file, "r") as f:
        source_content = f.read()
    
    rules_match = re.search(r"^rules:\n(.*?)(?=^[a-zA-Z]|\Z)", source_content, re.MULTILINE | re.DOTALL)
    if not rules_match:
        print("Error: Could not find rules section in source file")
        sys.exit(1)
    
    rules_content = rules_match.group(1)
    
    # Read the target file
    with open(target_file, "r") as f:
        target_content = f.read()
    
    # Replace the rules section in the target file
    new_content = re.sub(r"^rules:.*?(?=^[a-zA-Z]|\Z)", f"rules:\n{rules_content}", target_content, flags=re.MULTILINE | re.DOTALL)
    
    # Write back to the target file
    with open(target_file, "w") as f:
        f.write(new_content)
    
    print("Rules section successfully synced")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python sync_rules.py <source_file> <target_file>")
        sys.exit(1)
    
    source_file = sys.argv[1]
    target_file = sys.argv[2]
    sync_rules(source_file, target_file) 