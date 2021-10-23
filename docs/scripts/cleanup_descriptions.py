import json
from pathlib import Path

import ruamel.yaml

from build_md_docs import JSON_SCHEMAS_DIR, DESCRIPTIONS_DIR


def get_unused_descriptions(descriptions, json_schema):
    unused = []

    for uuid in descriptions:
        if is_uuid_used(uuid, json_schema) is False:
            unused.append(uuid)

    return unused


def load_descriptions():
    with open(DESCRIPTIONS_DIR / 'ru.yaml', 'r', encoding='utf-8') as descriptions_f:
        yaml = ruamel.yaml.YAML()
        descriptions = yaml.load(descriptions_f)
    return descriptions


def load_json_schema():
    schema_path = JSON_SCHEMAS_DIR / 'project.schema.json'
    with open(schema_path, 'r') as schema_f:
        whole_schema = json.load(schema_f, object_hook=assemble_json_schema(schema_path))

    return whole_schema


def assemble_json_schema(schema_path):
    work_dir = Path(schema_path).parent.resolve()

    def wrapped(obj):
        if '$ref' in obj:
            if '#/definitions/' not in obj['$ref']:
                ref_schema_path = work_dir / obj['$ref']
                with open(ref_schema_path, 'r') as schema_f:
                    obj = json.load(schema_f, object_hook=assemble_json_schema(ref_schema_path))
        return obj

    return wrapped


def is_uuid_used(uuid, var):
    if 'uuid' in var and var['uuid'] == uuid:
        return True

    for k, val in var.items():
        if isinstance(val, dict):
            if is_uuid_used(uuid, val):
                return True

        if isinstance(val, list):
            for i in val:
                if isinstance(i, dict):
                    if is_uuid_used(uuid, i):
                        return True

    return False


if __name__ == '__main__':
    descriptions = load_descriptions()
    json_schema = load_json_schema()
    unused = get_unused_descriptions(descriptions, json_schema)

    for uuid in unused:
        del descriptions[uuid]

    with open(DESCRIPTIONS_DIR / 'ru.yaml', 'w', encoding='utf-8') as descriptions_f:
        yaml = ruamel.yaml.YAML()
        data = yaml.dump(descriptions, descriptions_f)
