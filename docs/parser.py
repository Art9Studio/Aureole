class Md:
    @staticmethod
    def bold(value):
        return f'**{value}**'

    @staticmethod
    def code(value):
        return f'`{value}`'

    @staticmethod
    def list_item(value):
        return f'* {value}'

    @staticmethod
    def h1(value):
        return f'# {value}'

    @staticmethod
    def h2(value):
        return f'## {value}'

    @staticmethod
    def h3(value):
        return f'### {value}'

    @staticmethod
    def h4(value):
        return f'#### {value}'

    @staticmethod
    def hr():
        return '***'


class Parser:
    tab_size = 2

    @staticmethod
    def create_description(obj, obj_type):
        description = []

        if 'type' in obj:
            description.append(f"Тип: {Md.bold(obj['type'])}.")
        if 'const' in obj:
            description.append(f"Константа: {Md.bold(obj['const'])}.")
        if 'enum' in obj:
            description.append(f"Допускаются следующие значения: [{', '.join(map(Md.code, obj['enum']))}].")
        if 'metainfo' in obj:
            description.append(f"Дополнительно: {obj['metainfo']}.")
        if 'default' in obj:
            description.append(f"Значение по-умолчанию: {Md.bold(obj['default'])}.")

        if obj_type != 'general':
            if "title" in obj:
                description.append(obj['title'])
            if "description" in obj:
                description.append((obj['description']))

        if 'required' in obj:
            description.append(f"Обязательны: [{', '.join(map(Md.code, obj['required']))}].")
        if 'additionalProperties' in obj:
            val = 'Допустимо' if obj['additionalProperties'] else 'Недопустимо'
            description.append(f"Наличие дополнительных свойств: {Md.bold(val)}.")
        if 'additionalItems' in obj:
            val = 'Допустимо' if obj['additionalItems'] else 'Недопустимо'
            description.append(f"Наличие дополнительных элементов: {Md.bold(val)}.")
        if 'minimum' in obj:
            description.append(f"Минимальное значение: {Md.bold(obj['minimum'])}.")
        if 'minItems' in obj:
            description.append(f"Минимальное кол-во элементов: {Md.bold(obj['minItems'])}.")
        if 'minLength' in obj:
            description.append(f"Минимальная длина: {Md.bold(obj['minLength'])}.")
        if 'minProperties' in obj:
            description.append(f"Минимальное кол-во свойств: {Md.bold(obj['minProperties'])}.")
        if 'maximum' in obj:
            description.append(f"Максимальное значение: {Md.bold(obj['maximum'])}.")
        if 'maxItems' in obj:
            description.append(f"Максимальное кол-во элементов: {Md.bold(obj['maxItems'])}.")
        if 'maxLength' in obj:
            description.append(f"Максимальная длина: {Md.bold(obj['maxLength'])}.")
        if 'maxProperties' in obj:
            description.append(f"Максимальное кол-во свойств: {Md.bold(obj['maxProperties'])}.")
        if 'uniqueItems' in obj:
            if obj['uniqueItems']:
                description.append('Элементы должны быть уникальными.')
            else:
                description.append('Элементы могут быть неуникальными.')
        if '$ref' in obj:
            description.append(f"См. {obj['$ref']}.")

        return description

    def parse_object(self, obj, name, output_lines=None, indent_level=0, obj_type='inner'):
        if not output_lines:
            output_lines = []

        indentation = " " * self.tab_size * indent_level

        if obj_type == 'general':
            if "title" in obj:
                output_lines.append(indentation + Md.h1(obj['title']) + '\n')
                output_lines.append(Md.hr() + '\n')
            if "description" in obj:
                output_lines.append(indentation + obj['description'] + '\n')

        description_line = " ".join(self.create_description(obj, obj_type))

        if name == '':
            name = None
        elif name == 'items':
            name = Md.bold('Элементы')
        elif name == 'definitions':
            name = Md.bold('Определения')
        else:
            name = Md.code(name)

        if not name and description_line:
            output_lines.append(f"{indentation}- {description_line}\n")
        elif name and description_line:
            output_lines.append(f"{indentation}- {name}: {description_line}\n")

        if 'items' in obj:
            output_lines = self.parse_object(obj['items'], 'items', output_lines=output_lines,
                                             indent_level=indent_level + 1)

        if "properties" in obj:
            for property_name, property_obj in obj["properties"].items():
                output_lines = self.parse_object(property_obj, property_name, output_lines=output_lines,
                                                 indent_level=indent_level + 1)

        if "patternProperties" in obj:
            for property_name, property_obj in obj["patternProperties"].items():
                output_lines = self.parse_object(property_obj, f'regex("{property_name}")', output_lines=output_lines,
                                                 indent_level=indent_level + 1)

        if "oneOf" in obj:
            i = " " * self.tab_size * (indent_level + 1)
            output_lines.append(i + Md.h2('Допустим один из вариантов конфига:') + '\n')
            for property_obj in obj["oneOf"]:
                if 'title' in property_obj:
                    output_lines.append(i + Md.h3(property_obj['title']) + '\n')

                output_lines = self.parse_object(property_obj, "", output_lines=output_lines,
                                                 indent_level=indent_level + 1)

        if "anyOf" in obj:
            i = " " * self.tab_size * (indent_level + 1)
            output_lines.append(i + Md.h3('Допустим любой из вариантов конфига:') + '\n')
            for property_obj in obj["anyOf"]:
                output_lines.append(i + Md.h4('{') + '\n')
                output_lines = self.parse_object(property_obj, "", output_lines=output_lines,
                                                 indent_level=indent_level + 1)
                output_lines.append(i + Md.h4('}') + '\n')

        if "definitions" in obj:
            output_lines.append(indentation + Md.h3('Определения:') + '\n')
            for def_name, def_obj in obj["definitions"].items():
                output_lines.append(indentation + Md.h3(def_name) + '\n')
                output_lines = self.parse_object(def_obj, def_name, output_lines=output_lines)

        return output_lines

    def parse_schema(self, schema):
        output_lines = self.parse_object(schema, '', obj_type='general')
        return output_lines
