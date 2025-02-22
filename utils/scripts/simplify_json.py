#!/usr/bin/env python3
import json
import sys
from pathlib import Path

def simplify_schema(data):
    if isinstance(data, dict):
        return {k: simplify_schema(v) for k, v in data.items()}
    elif isinstance(data, list):
        return [simplify_schema(data[0])] if data else []
    elif isinstance(data, (int, float)):
        return 0
    elif isinstance(data, bool):
        return False
    elif isinstance(data, str):
        return ""
    return None

def process_json_file(file_path: str):
    with open(file_path, 'r') as f:
        data = json.load(f)
    return simplify_schema(data)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Uso: ./script.py <ruta_del_json>")
        sys.exit(1)
        
    json_path = sys.argv[1]
    try:
        result = process_json_file(json_path)
        print(json.dumps(result, indent=2))
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


#  python ./utils/scripts/simplify_json.py /home/devel/chivomap/api/utils/assets/topo.json
# {
#   "type": "",
#   "objects": {
#     "collection": {
#       "type": "",
#       "geometries": [
#         {
#           "type": "",
#           "arcs": [
#             [
#               [
#                 0
#               ]
#             ]
#           ],
#           "properties": {
#             "NAM": "",
#             "D": "",
#             "M": ""
#           }
#         }
#       ]
#     }
#   },
#   "arcs": [
#     [
#       [
#         0
#       ]
#     ]
#   ],
#   "bbox": [
#     0
#   ]
# }