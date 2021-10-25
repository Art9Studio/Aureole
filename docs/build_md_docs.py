import json
import operator
from functools import reduce
from pathlib import Path

import yaml

from md_parser import MDParser

BASE_DIR = Path(__file__).parent.resolve()
JSON_SCHEMAS_DIR = BASE_DIR / 'json_schemas'
DESCRIPTIONS_DIR = BASE_DIR / 'descriptions'
EXAMPLES_DIR = BASE_DIR / 'config_examples'
MD_DIR = BASE_DIR / 'vuepress_docs/docs/config'


def load_json_schema():
    schema_path = JSON_SCHEMAS_DIR / 'project.schema.json'
    with open(schema_path, 'r') as schema_f:
        whole_schema = json.load(schema_f, object_hook=assemble_json_schema(schema_path))

    return whole_schema


def assemble_json_schema(schema_path):
    work_dir = Path(schema_path).parent.resolve()
    descriptions = load_descriptions()
    metatags = load_metatags()
    examples = load_examples()

    def wrapped(obj):
        if 'uuid' in obj:
            description = descriptions[obj['uuid']]
            if description['title']:
                obj['title'] = description['title']
            if description['description']:
                obj['description'] = description['description']

        if 'metatags' in obj:
            info = ''
            for tag in obj['metatags'].split(' '):
                if tag.find(':') != -1:
                    pure_tag, additional = tag.split(':')
                    info += metatags[pure_tag] + ' ' + additional + ', '
                else:
                    info += metatags[tag] + ', '
            obj['metainfo'] = info.rstrip(', ')

        if 'example' in obj:
            obj['example'] = examples[obj['example']]

        if '$ref' in obj:
            ref = obj['$ref']
            if '#/definitions/' in ref:
                definition_path, definition_name = ref.split('#/definitions/')
                definition_ref = definition_name.lower().replace(' ', '-')

                if definition_path:
                    plugin_names = [d.name for d in Path('../plugins').iterdir() if d.is_dir()] + ['collection',
                                                                                                      'identity']
                    plugin_name = list(set(plugin_names) & set(p.name for p in Path(definition_path).parents))[0]
                    obj['$ref'] = f'[{definition_name}](./{plugin_name}.md#{definition_ref})'
                else:
                    obj['$ref'] = f'[{definition_name}](#{definition_ref})'
            else:
                ref_schema_path = work_dir / ref
                with open(ref_schema_path, 'r') as schema_f:
                    obj = json.load(schema_f, object_hook=assemble_json_schema(ref_schema_path))
        return obj

    return wrapped


def load_descriptions():
    with open(DESCRIPTIONS_DIR / 'ru.yaml', 'r', encoding='utf-8') as descriptions_f:
        descriptions = yaml.load(descriptions_f, Loader=yaml.BaseLoader)
    return descriptions


def load_metatags():
    with open(JSON_SCHEMAS_DIR / 'metatags.yaml', 'r', encoding='utf-8') as metatags_f:
        metatags = yaml.load(metatags_f, Loader=yaml.BaseLoader)['metatags']

    with open(DESCRIPTIONS_DIR / 'ru_meta.yaml', 'r', encoding='utf-8') as descriptions_f:
        meta_descriptions = yaml.load(descriptions_f, Loader=yaml.BaseLoader)

    for tag, uuid in metatags.items():
        metatags[tag] = meta_descriptions[uuid]

    return metatags


def load_examples():
    examples = {}

    path = Path(EXAMPLES_DIR)
    for p in path.glob("**/*.yaml"):
        rel_path = p.relative_to(path)
        with open(EXAMPLES_DIR / rel_path) as examples_f:
            key = (rel_path.parent / rel_path.stem).as_posix()
            examples[key] = examples_f.read()

    return examples


def split_json_schema(project_schema):
    props_keys = ['properties']
    app_props_keys = props_keys + ['apps', 'items', 'properties']
    plugin_keys = {
        'authn': app_props_keys + ['authN', 'items'],
        'authz': app_props_keys + ['authZ', 'items'],
        'identity': app_props_keys + ['identity'],
        'storage': props_keys + ['storages', 'items'],
        'collection': props_keys + ['collections', 'items'],
        'hasher': props_keys + ['hashers', 'items'],
        'crypto_key': props_keys + ['crypto_keys', 'items'],
        'sender': props_keys + ['senders', 'items'],
        'admin_plugin': props_keys + ['admin_plugins', 'items']
    }

    plugin_schemas = {}
    for plugin_name, keys in plugin_keys.items():
        plugin_schemas[plugin_name] = get_by_path(project_schema, keys)
        pretty_name = ''.join(word.title() for word in plugin_name.split('_'))
        ref = f'[{pretty_name}](./{plugin_name}.md)'
        set_by_path(project_schema, keys, {'$ref': ref})

    return plugin_schemas


def get_by_path(root, keys):
    return reduce(operator.getitem, keys, root)


def set_by_path(root, keys, value):
    get_by_path(root, keys[:-1])[keys[-1]] = value


if __name__ == '__main__':
    project_schema = load_json_schema()
    plugin_schemas = split_json_schema(project_schema)

    parser = MDParser()
    project_md = parser.parse_schema(project_schema)
    with open(MD_DIR / 'project.md', 'w', encoding='utf-8') as project_md_f:
        project_md_f.writelines(project_md)

    for plugin_name, plugin_schema in plugin_schemas.items():
        plugin_md = parser.parse_schema(plugin_schema)

        if plugin_name == 'collection':
            plugin_md.insert(0, '---\nsidebarDepth: 0\n---\n')

        with open(MD_DIR / f'{plugin_name}.md', 'w', encoding='utf-8') as plugin_md_f:
            plugin_md_f.writelines(plugin_md)
