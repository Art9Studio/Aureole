# Identity менеджер
***
Описывает все свойства данного identity менеджера. Плагин берет на себя роль взаимодействия с БД и предоставляет ряд методов для получения и сохранения данных пользователя.
- Тип: **object**. Наличие дополнительных свойств: **Недопустимо**.
  #### Допустим один из вариантов конфига
  ## Конфигурация jwt webhook identity менеджера.
  - Тип: **object**. Конфигурация jwt webhook identity менеджера. Служит для задания свойств менеджера. Обязательны: `type`, `config`. Наличие дополнительных свойств: **Недопустимо**.
    - `type`: Константа: **jwt_webhook**. Тип менеджера. Необходим, чтобы дать Aureole понять, какой из менеджеров использовать.
    - `config`: Тип: **object**. Конфигурация identity менеджера. Описывает все свойства данного менеджера. Обязательны: `address`. Наличие дополнительных свойств: **Недопустимо**.
      - `address`: Тип: **string**. Дополнительно: Абсолютный URL-адрес. Адрес сервера. Хост и порт сервера, к которому будет обращаться плагин.
      - `retries_num`: Тип: **integer**. Количество повторений запроса. Максимальное количество повторений запроса, которое будет совершать плагин в случае возникновения ошибки. Минимальное значение: **1**.
      - `retry_interval`: Тип: **integer**. Дополнительно: Единицы измерения: ms. Интервал между запросами. Время, которое будет ждать плагин, перед тем, как совершить очередную попытку сделать запрос. Минимальное значение: **1**.
      - `timeout`: Тип: **integer**. Дополнительно: Единицы измерения: ms. Время ожидания ответа от сервера. Максимальное время ожидание ответа, по истечению которого, плагин перестанет ждать ответа. Минимальное значение: **1**.
      - `headers`: Тип: **object**. Перечисление заголовков. Заголовки, которые будут приложены к запросу.
        - `.*`: Тип: **string**.
    ### Пример конфига
    ```yaml
    id_managers:
      - type: "jwt_webhook"
        name: webhook_identity
        config:
          address: http://localhost:3001
          retries_num: 1
          retry_interval: 2
          timeout: 10
          headers:
            header1: value1
            header2: value2
    ```
  ## Конфигурация standard менеджера.
  - Тип: **object**. Конфигурация standard менеджера. Служит для задания свойств менеджера. Обязательны: `type`, `config`. Наличие дополнительных свойств: **Недопустимо**.
    - `type`: Константа: **standard**. Тип менеджера. Необходим, чтобы дать Aureole понять, какой из менеджеров использовать.
    - `config`: Тип: **object**. Конфигурация identity менеджера. Описывает все свойства данного менеджера. Обязательны: `db_url`. Наличие дополнительных свойств: **Недопустимо**.
      - `db_url`: Тип: **string**. URL для подклчючения к БД. По этому URl-у плагин будет совершать подключения к БД. Минимальная длина: **1**.
    ### Пример конфига
    ```yaml
    id_managers:
      - type: "standard"
        name: standard
        config:
          db_url: postgresql://root:password@localhost:5432/aureole
    ```