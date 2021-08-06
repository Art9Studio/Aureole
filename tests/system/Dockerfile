FROM python:3.9.1

ENV PYTHONUNBUFFERED 1

WORKDIR /app

COPY requirements.txt requirements.txt

RUN pip install --no-cache-dir -r requirements.txt

WORKDIR tests/

EXPOSE 8000

ENTRYPOINT system/entrypoint.sh