from django.contrib.auth.base_user import AbstractBaseUser
from django.contrib.auth.models import UserManager
from django.db import models


class CustomUser(AbstractBaseUser):
    username = models.TextField(unique=True)
    phone = models.TextField(unique=True, null=True)
    email = models.EmailField(unique=True, null=True)
    password = models.TextField()

    USERNAME_FIELD = 'username'

    objects = UserManager()

    class Meta:
        db_table = "users"
