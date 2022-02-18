
<p align="center">
<img src="assets/Logo.svg" width="300px"/>
 </p>
<p align="center" color="#1FD6E3">Самый гибкий и современный сервер аутентификации </br>с открытым исходным кодом.</p>

<p align="center">
<img src="https://img.shields.io/github/commit-activity/m/Art9Studio/Aureole">
<a href="https://discord.gg/EjBQ3fKg"><img src="https://img.shields.io/badge/chat-discord-brightgreen.svg?logo=discord&style=flat"></a>
<a href="https://twitter.com/aureolecloud"><img src="https://img.shields.io/badge/Follow-aureolecloud-blue.svg?style=flat&logo=twitter"></a>
</p>

# 🔥 О проекте

**Aureole** – сервер аутентификации и управления пользователями с открытым исходным кодом, быстрой интеграцией с любым стеком на вашем проекте, а также с модульной масштабируемой архитектурой и встроенным набором плагинов.

<!--**Aureole** предоставляет совокупность самых важных функций &quot;из коробки&quot;: нативная поддержка **Hasura, PostgresSQL, Django,** популярные способы аутентификации **Google, Apple ID, Facebook, с подтверждением по смс и email**. Если этого окажется недостаточно, то наша архитектура даст возможность быстро написать новый плагин под ваши бизнес процессы.-->
<!--<img src="https://github.com/savkovbohdan/ViFit/blob/master/GifVideo.png" width="500px"/>-->

# 📍Статус

- [x] Pre-Alpha: Разработка и тестирование базового набора плагинов
- [ ] Alpha: Исправление багов, покрытие тестами
- [ ] Beta: Запуск решения для закрытой группы клиентов
- [ ] Release candidate: Открытое тестирование

Сейчас мы находимся на **ранней версии продукта (Pre-Alpha)**. Чтобы получать самые свежие версии сборок следите за обновлениями нашего репозитория **(ветка main)**.

# ⚡Фичи

- Гибкая архитектура благодаря системе плагинов
- Обширный набор плагинов аутентификации
- Работает с JWT
- Может настраиваться под ваши бизнес процессы и решения (из коробки есть пресеты для Hasura, Django)
- Поддерживает вашу БД при наличии соответствующего плагина
- Независимость от языка вашей системы
- Набор методов хеширования, криптографических алгоритмов

# 📖 Содержание ![](RackMultipart20210814-4-ec3q2v_html_cb55ddb5edd60516.gif)

- [Быстрый запуск:](#-быстрый-запуск-)
    - [Развертывание в один клик](#развертывание-в-один-клик)
- [Архитектура](#-архитектура)
- [Бизнес кейсы](#бизнес-кейсы)
- [Плагины](#%EF%B8%8F-плагины)
    - [О плагинах](#о-плагинах)
    - [Аутентификационные плагины](#аутентификационные-плагины)
    - [Плагины для двухфакторной аутентификации](#плагины-для-двухфакторной-аутентификации)
    - [Авторизационные плагины](#авторизационные-плагины)
    - [Плагины для хранилища](#плагины-для-хранилища)
    - [Плагины для хеширования](#плагины-для-хеширования)
    - [Плагины для импорта ключей](#плагины-для-импорта-ключей)
    - [Плагины для отправки сообщений](#плагины-для-отправки-сообщений)
- [Поддержка и устранение багов](#-поддержка-и-устранение-багов)
- [Оценки](#-оценили)
- [Сделали Fork](#%EF%B8%8F-сделали-fork)
- [Лицензия](#-лицензия)
- [Переводы](#%EF%B8%8F-переводы)

# 🚀 Быстрый запуск: ![](RackMultipart20210814-4-ec3q2v_html_cb55ddb5edd60516.gif)


## Развертывание в один клик:

| Провайдер | Ссылка | Документация |
| --- | --- | --- |
| Heroku | [![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy) | Ссылка |
| Render | [![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/Art9Studio/Aureole) | Ссылка |

# ⚙ Архитектура

Aureole легко встраивается в любую архитектуру вашего продукта. Он будет принимать запросы на аутентификацию и выдавать JWT или Cookie-сессии.
Также он отвечает за регистрацию, смену и восстановление паролей, хеширование паролей в БД.

 <img src="assets/Scheme.svg" width="500px"/>

# 👾Бизнес кейсы

- Регистрация и авторизация пользователей на вашем сайте
- Замена стандартного механизма аутентификации в Django
- Аутентификация ваших пользователей в Hasura
- Единая аутентификация между доменами
- Аутентификация в Kubernetes через OpenID
- Аутентификация в Docker Registry
- Аутентификация embedded устройств

# 🖇️ Плагины

Мы разработали базовый набор плагинов и предоставляем возможность удобно расширить функциональность существующих плагинов и быстро разработать новые  под ваши бизнес кейсы.

К концу года планируется запустить магазин плагинов.

## Аутентификационные плагины

- [x] Логин-пароль
- [x] Passwordless по E-mail
- [x] Passwordless по SMS
- [x] Google OAuth 2.0
- [x] Facebook OAuth 2.0
- [x] VK OAuth 2.0
- [x] Apple ID
- [ ] GitHub
- [ ] Instagram
- [ ] Challenge-response authentication

## Плагины для двухфакторной аутентификации

- [x] SMS
- [x] Google Authenticator
- [ ] YubiKey

## Авторизационные плагины

- [x] JWT

## Плагины для хранилища

- [x] PostgreSQL
- [ ] MongoDB
- [ ] MySQL

## Плагины для хеширования

- [x] Argon2
- [x] Pbkdf2 (Django)

## Плагины для импорта ключей

- [x] JWK
- [x] Pem

## Плагины для отправки сообщений

- [x] E-mail (SMTP)
- [x] Twillio

# 💬 Поддержка и устранение багов

Документация и сообщество поможет вам решить любую проблему. Если вы столкнулись с ошибкой или вам нужно связаться с нами, вы можете использовать один из следующих каналов связи:

- Поддержка и обратная связь: [Discord](https://discord.gg/EjBQ3fKg)
- Проблема и отслеживание ошибок: [GitHub issues](https://github.com/Art9Studio/Aureole/issues)
- Следите за обновлениями продукта в Twitter: [@aureolecloud](https://twitter.com/aureolecloud)
- Поговорите с нами в чате: [Telegram](https://t.me/joinchat/lsaDf65QlHk5M2Ri)
- Написать на E-mail: [hi@aureole.cloud](mailto:hi@aureole.cloud)

# ⭐ Оценили
[![Stargazers repo roster for @USERNAME/REPO_NAME](https://reporoster.com/stars/Art9Studio/Aureole)](https://github.com/Art9Studio/Aureole/stargazers)

# 🛠️ Сделали Fork
[![Forkers repo roster for @USERNAME/REPO_NAME](https://reporoster.com/forks/Art9Studio/Aureole)](https://github.com/Art9Studio/Aureole/network/members)


# 📝 Лицензия 

The core Aureole is available under the [GNU Affero General Public
License v3](https://www.gnu.org/licenses/agpl-3.0.en.html) (AGPL-3.0).

**Commercial licenses** are available on request, if you do not wish to use the Aureole under the AGPL license. Typically, they come bundled with support plans and SLAs. Please feel free to contact us at [hi@aureole.cloud](mailto:hi@aureole.cloud).

All **other contents** (except those in [`internal`](internal) and
[`plugins`](plugins) directories) are available under the [MIT License](LICENSE-community).
This includes everything in all other directories.

# 🈂️ Переводы

- [English 🇬🇧](https://github.com/Art9Studio/Aureole)
- [Russian 🇷🇺](https://github.com/Art9Studio/Aureole)
