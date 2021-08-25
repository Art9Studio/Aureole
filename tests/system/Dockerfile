FROM python:3.9.1

ENV PYTHONUNBUFFERED 1

WORKDIR /app/tests

COPY requirements.txt requirements.txt

RUN pip install --no-cache-dir -r requirements.txt

ENTRYPOINT system/entrypoint.sh