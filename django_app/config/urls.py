from django.contrib import admin
from django.urls import include, path

from core.views import index, user_info

urlpatterns = [
    path("admin/", admin.site.urls),
    path("oidc/", include("mozilla_django_oidc.urls")),
    path("api/", include("core.api_urls")),
    path("userinfo/", user_info, name="userinfo"),
    path("", index, name="index"),
]

