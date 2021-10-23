# Хранилище
***
Описывает все свойства данного хранилища. Данный плагин используется для работы с различными хранилищами данных. В плагине описаны методы, которые используется для сохранения, изменения и удаления данных. Для работы плагина требуется существующее хранилище.
- Тип: **object**. Обязательны: `name`. Наличие дополнительных свойств: **Недопустимо**.
  - `name`: Тип: **string**. Дополнительно: Имя экземпляра плагина. Имя хранилища. Имя используется для того, чтобы в дальнейшем ссылаться на данное хранилище. Минимальная длина: **1**.
  #### Допустим один из вариантов конфига
  ## Конфигурация postgresql хранилища.
  - Тип: **object**. Конфигурация postgresql хранилища. Служит для задания свойств хранилища. Обязательны: `type`, `config`. Наличие дополнительных свойств: **Недопустимо**.
    - `type`: Константа: **postgresql**. Тип хранилища. Необходим, чтобы дать Aureole понять, какое из хранилищ использовать.
    - `config`: Конфигурация хранилища. Описывает все свойства данного хранилища.
      #### Допустим один из вариантов конфига
      - Тип: **object**. Обязательны: `url`. Наличие дополнительных свойств: **Недопустимо**.
        - `url`: Тип: **string**. Дополнительно: Абсолютный URL-адрес. Строка подключения к базе. Строка подключения к базе в формате DSN (Data Source Name).
      - Тип: **object**. Обязательны: `username`, `password`, `host`, `port`, `db_name`. Наличие дополнительных свойств: **Недопустимо**.
        - `username`: Тип: **string**. Имя пользователя. Имя пользователя для аутентификации в БД. Минимальная длина: **1**.
        - `password`: Тип: **string**. Пароль пользователя. Пароль пользователя для аутентификации в БД. Минимальная длина: **1**.
        - `host`: Тип: **string**. Хост БД. Наименование хоста, на котором работает БД. Минимальная длина: **1**.
        - `port`: Тип: **string**. Порт БД. Наименование порта, на котором работает БД. Минимальная длина: **1**.
        - `db_name`: Тип: **string**. Имя БД. Имя БД, к которой необходимо подключиться. Минимальная длина: **1**.
        - `options`: Тип: **object**. Перечисление опций. Список дополнительный опций.
          - `regex(".+")`: Тип: **string**.
    ### Пример конфига
    ```yaml
    storages:
      - type: "postgresql"
        name: db
        config:
          url: "postgresql://root:password@localhost:5432/aureole?sslmode=disable&search_path=public"
    
      - type: "postgresql"
        name: db_two
        config:
          username: "root"
          password: "password"
          host: "localhost"
          port: "5432"
          db_name: "aureole"
          options:
            sslmode: "disable"
            search_path: "public"
    
    ```