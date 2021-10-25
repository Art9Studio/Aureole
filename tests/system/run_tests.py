import json
import os
import pathlib
import pprint
import re
import subprocess
import uuid
from pathlib import Path

import jinja2
import yaml
from yaml import safe_dump
from yamlreader import data_merge

WORK_DIR = Path(__file__).parent.resolve()
RES_DIR = Path(__file__).parent.resolve() / 'resources'
CLEANUP_FILES_QUEUE = []
matcher = re.compile(r"\$\{([a-zA-Z_$0-9]+)(:-.*)?\}")


def prepare_tests():
    modules = get_modules()
    populate_modules_vars(modules)
    pprint.pprint(modules)
    ensure_resources_dir(modules)
    reveal_uuids_to_modules(modules)

    prepare_psql_dbs(modules)
    prepare_psql_schemas(modules)
    merge_aureole_configs(modules)
    prepare_module_resources(modules)


def finish_tests():
    clean_up_docker()
    clean_up_files()


def prepare_module_resources(modules):
    tmpl_path = os.path.join(RES_DIR, 'docker-compose.yml.j2')
    loader = jinja2.FileSystemLoader(searchpath="/")
    env = jinja2.Environment(loader=loader)

    target_path = os.path.join(RES_DIR, 'docker-compose.yml')
    with open(target_path, 'w') as f:
        f.write(env.get_template(tmpl_path).render(modules=modules))

    CLEANUP_FILES_QUEUE.append(target_path)


def prepare_psql_dbs(modules):
    psql_dbs_path = os.path.join(RES_DIR, 'psql-dbs.sql')
    with open(psql_dbs_path, 'w') as f:
        for module in modules:
            module_uuid = module.get('uuid')
            sql = f'''CREATE DATABASE "{module_uuid}"; GRANT ALL PRIVILEGES ON DATABASE "{module_uuid}" TO root;\n'''
            f.write(sql)

    CLEANUP_FILES_QUEUE.append(psql_dbs_path)


def prepare_psql_schemas(modules):
    psql_schemas_path = os.path.join(RES_DIR, 'psql-schemas.sql')
    with open(psql_schemas_path, 'w') as f:
        for module in modules:
            schema_path = os.path.join(module.get('path'), 'resources', 'schema.sql')
            if os.path.isfile(schema_path):
                with open(schema_path, 'r') as schema_f:
                    schema = schema_f.readlines()

                module_uuid = module.get('uuid')
                sql = f'\\connect "{module_uuid}"'
                f.write(sql + "\n")
                f.writelines(schema)

    CLEANUP_FILES_QUEUE.append(psql_schemas_path)


def ensure_resources_dir(modules):
    for module in modules:
        res_dir_path = os.path.join(module.get('path'), "resources")
        Path(res_dir_path).mkdir(parents=True, exist_ok=True)


def reveal_uuids_to_modules(modules):
    module_uuids = {}
    for module in modules:
        module_uuids[module.get('name')] = module.get('uuid')

    target_path = os.path.join(RES_DIR, "uuid.txt")
    with open(target_path, 'w') as f:
        f.write(json.dumps(module_uuids))

    CLEANUP_FILES_QUEUE.append(target_path)


def populate_modules_vars(modules):
    for module in modules:
        module.update(get_db_connection(module))
        module.update(get_aureole_resource_pathes(module))


def get_modules():
    modules = []
    pathes = []
    for path in Path(Path(__file__).parent.resolve()).iterdir():
        if path.is_dir() and path.name.startswith('test_'):
            if path not in pathes:
                modules.append({'path': str(path), 'uuid': str(uuid.uuid4()), 'name': path.name})
                pathes.append(path)
    return modules


def merge_aureole_configs(modules):
    configs = []
    for module in modules:
        conf_path = os.path.join(module.get('path'), "resources", "config.yaml")
        if Path(conf_path).is_file():
            configs.append({'path': conf_path, 'module': module})
    merge_yamls(configs)


def clean_up_files():
    for file in CLEANUP_FILES_QUEUE:
        pathlib.Path(file).unlink()


def clean_up_docker():
    subprocess.run('cd tests/system/resources && docker-compose down',
                   shell=True, check=True, text=True)


#############


# Helpers
#############
def get_aureole_resource_pathes(module):
    return {'res_path': f'/resources/{module.get("uuid")}'}


def get_db_connection(module):
    prefix = 'db_connection_'
    return {prefix + 'psql': f'postgresql://root:password@postgres:5432/{module.get("uuid")}'}


def merge_yamls(configs):
    data = {}
    for config in configs:
        path = config.get('path')
        with open(path, 'r') as f:
            text = f.read()

            text = interpolate_vars(text, config.get('module'))
            new_data = yaml.load(text, Loader=yaml.FullLoader)
        if new_data is not None:
            data = data_merge(data, new_data)

    path = os.path.join(RES_DIR, 'config.yaml')
    with open(path, 'w') as f:
        f.write(safe_dump(data, indent=2, default_flow_style=False, canonical=False))

    CLEANUP_FILES_QUEUE.append(path)


def interpolate_vars(text, module):
    def repl(match):
        variable, default = match.groups()  # type: ignore

        if default:
            # lstrip() is dangerous!
            # It can remove legitimate first two letters in a value starting with `:-`
            default = default[2:]

        return module.get(variable.lower(), default)

    text = re.sub(matcher, repl, text)
    return text


#############


def run_and_wait_docker():
    try:
        subprocess.run(
            'cd tests/system/resources && docker-compose up -d aureole postgres smtp twilio social-auth && docker-compose up aureole-tests',
            shell=True, check=True, text=True)
    except:
        return


if __name__ == '__main__':
    prepare_tests()
    run_and_wait_docker()
    finish_tests()
