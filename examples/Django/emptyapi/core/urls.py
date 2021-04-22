from django.urls import path
from .views import index

app_name = "articles"

urlpatterns = [
    path('index/', index),
]