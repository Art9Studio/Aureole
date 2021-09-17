# Aureole Config
***
- Тип: **object**. Обязательны: [`api_version`, `apps`, `collections`, `storages`]. Наличие дополнительных свойств: **Недопустимо**.
  - `api_version`: Тип: **string**. Версия API. Версия API указывается для обратной совместимости при грядущих обновлениях продукта.
  - `ping_path`: Тип: **string**. Дополнительно: Относительный URL-адрес. Значение по-умолчанию: **/ping**. Маршрут для пинга сервера. Маршрут для пинга используется, чтобы понять, находится ли сервер в данный момент в рабочем состоянии.
  - `apps`: Тип: **array**. Перечисление приложений. Используется для указания конфигураций всех приложений, которые использует Aureole. Минимальное кол-во элементов: **1**.
    - **Элементы**: Тип: **object**. Конфигурация одного приложения. Используется для конфигурирования одного приложения. Обязательны: [`name`, `host`, `path_prefix`, `identity`, `authN`, `authZ`]. Наличие дополнительных свойств: **Недопустимо**.
      - `name`: Тип: **string**. Имя приложения. Используется для идентификации конфигурации определенного приложения. Минимальная длина: **1**.
      - `host`: Тип: **string**. Дополнительно: Абсолютный URL-адрес. Значение по-умолчанию: **http://localhost:3000**. Хост, на котором работает приложение. Указывается в формате <протокол>://<хост>:<порт>. Если не указан протокол, то будет использован HTTP. Хост используется для формирования абсолютных маршрутов приложения.
      - `path_prefix`: Тип: **string**. Дополнительно: Относительный URL-адрес. Значение по-умолчанию: ****. Префикс маршрута приложения. Префикс используется для образования пространства имен данного приложения. Должен быть уникальным в пределах перечисления.
      - `identity`: См. [Identity](./identity.md).
      - `authN`: Тип: **array**. Перечисление конфигураций аутентификаторов. Используется для указания конфигураций всех аутентификаторов, которые используются в приложении. Каждый аутентификатор независим от других и должен использовать свои, уникальные в пределах приложения, маршруты. Минимальное кол-во элементов: **1**.
        - **Элементы**: См. [Authn](./authn.md).
      - `authZ`: Тип: **array**. Перечисление конфигураций авторизаторов. Используется для указания конфигураций всех авторизаторов, которые используются в приложении. Каждый авторизатор независим от других и должен использовать свои, уникальные в пределах приложения, маршруты. Минимальное кол-во элементов: **1**.
        - **Элементы**: См. [Authz](./authz.md).
  - `storages`: Тип: **array**. Перечисление конфигураций хранилищ. Используется для указания конфигураций всех хранилищ, которые используются в проекте. Минимальное кол-во элементов: **1**.
    - **Элементы**: См. [Storage](./storage.md).
  - `collections`: Тип: **array**. Перечисление конфигураций коллекций. Используется для указания конфигураций всех коллекций, которые используются в проекте. Минимальное кол-во элементов: **1**.
    - **Элементы**: См. [Collection](./collection.md).
  - `hashers`: Тип: **array**. Перечисление конфигураций хэшеров. Используется для указания конфигураций всех хэшеров, которые используются в проекте.
    - **Элементы**: См. [Hasher](./hasher.md).
  - `crypto_keys`: Тип: **array**. Перечисление конфигураций ключей. Используется для указания конфигураций всех ключей, которые используются в проекте.
    - **Элементы**: См. [Crypto_key](./crypto_key.md).
  - `senders`: Тип: **array**. Перечисление конфигураций отправителей. Используется для указания конфигураций всех отправителей, которые используются в проекте.
    - **Элементы**: См. [Sender](./sender.md).