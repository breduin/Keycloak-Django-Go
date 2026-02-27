import os
from pathlib import Path

from django.core.management.utils import get_random_secret_key

BASE_DIR = Path(__file__).resolve().parent.parent

SECRET_KEY = os.getenv("DJANGO_SECRET_KEY", get_random_secret_key())
DEBUG = os.getenv("DJANGO_DEBUG", "0") == "1"

ALLOWED_HOSTS = os.getenv("ALLOWED_HOSTS", "*").split(",")

INSTALLED_APPS = [
    "django.contrib.admin",
    "django.contrib.auth",
    "django.contrib.contenttypes",
    "django.contrib.sessions",
    "django.contrib.messages",
    "django.contrib.staticfiles",
    "rest_framework",
    "mozilla_django_oidc",
    "core",
]

MIDDLEWARE = [
    "django.middleware.security.SecurityMiddleware",
    "django.contrib.sessions.middleware.SessionMiddleware",
    "django.middleware.common.CommonMiddleware",
    "django.middleware.csrf.CsrfViewMiddleware",
    "django.contrib.auth.middleware.AuthenticationMiddleware",
    "django.contrib.messages.middleware.MessageMiddleware",
    "django.middleware.clickjacking.XFrameOptionsMiddleware",
]

ROOT_URLCONF = "config.urls"

TEMPLATES = [
    {
        "BACKEND": "django.template.backends.django.DjangoTemplates",
        "DIRS": [],
        "APP_DIRS": True,
        "OPTIONS": {
            "context_processors": [
                "django.template.context_processors.debug",
                "django.template.context_processors.request",
                "django.contrib.auth.context_processors.auth",
                "django.contrib.messages.context_processors.messages",
            ],
        },
    },
]

WSGI_APPLICATION = "config.wsgi.application"

DATABASES = {
    "default": {
        "ENGINE": "django.db.backends.postgresql",
        "NAME": os.getenv("POSTGRES_DB", "djangoapp"),
        "USER": os.getenv("POSTGRES_USER", "django"),
        "PASSWORD": os.getenv("POSTGRES_PASSWORD", "django"),
        "HOST": os.getenv("POSTGRES_HOST", "django-db"),
        "PORT": os.getenv("POSTGRES_PORT", "5432"),
    }
}

AUTH_PASSWORD_VALIDATORS = [
    {
        "NAME": "django.contrib.auth.password_validation.UserAttributeSimilarityValidator",
    },
    {
        "NAME": "django.contrib.auth.password_validation.MinimumLengthValidator",
    },
    {
        "NAME": "django.contrib.auth.password_validation.CommonPasswordValidator",
    },
    {
        "NAME": "django.contrib.auth.password_validation.NumericPasswordValidator",
    },
]

LANGUAGE_CODE = "ru-ru"
TIME_ZONE = "UTC"
USE_I18N = True
USE_TZ = True

STATIC_URL = "static/"
STATIC_ROOT = BASE_DIR / "staticfiles"

DEFAULT_AUTO_FIELD = "django.db.models.BigAutoField"

AUTH_USER_MODEL = "core.CustomUser"

REST_FRAMEWORK = {
    "DEFAULT_PERMISSION_CLASSES": [
        "rest_framework.permissions.IsAuthenticated",
    ],
}

AUTHENTICATION_BACKENDS = (
    "core.oidc.KeycloakOIDCBackend",
    "django.contrib.auth.backends.ModelBackend",
)

LOGIN_URL = "oidc_authentication_init"
LOGIN_REDIRECT_URL = "/"
LOGOUT_REDIRECT_URL = "/"

OIDC_RP_CLIENT_ID = os.getenv("OIDC_RP_CLIENT_ID", "")
OIDC_RP_CLIENT_SECRET = os.getenv("OIDC_RP_CLIENT_SECRET", "")
OIDC_OP_ISSUER = os.getenv("OIDC_OP_ISSUER", "")
OIDC_PUBLIC_OP_ISSUER = os.getenv("OIDC_PUBLIC_OP_ISSUER", OIDC_OP_ISSUER)
OIDC_RP_SCOPES = "openid profile email"
OIDC_RP_SIGN_ALGO = "RS256"

if OIDC_OP_ISSUER:
    _OIDC_ISSUER_STRIPPED = OIDC_OP_ISSUER.rstrip("/")
    OIDC_OP_TOKEN_ENDPOINT = f"{_OIDC_ISSUER_STRIPPED}/protocol/openid-connect/token"
    OIDC_OP_USER_ENDPOINT = f"{_OIDC_ISSUER_STRIPPED}/protocol/openid-connect/userinfo"
    OIDC_OP_JWKS_ENDPOINT = f"{_OIDC_ISSUER_STRIPPED}/protocol/openid-connect/certs"

if OIDC_PUBLIC_OP_ISSUER:
    _OIDC_PUBLIC_ISSUER_STRIPPED = OIDC_PUBLIC_OP_ISSUER.rstrip("/")
    OIDC_OP_AUTHORIZATION_ENDPOINT = (
        f"{_OIDC_PUBLIC_ISSUER_STRIPPED}/protocol/openid-connect/auth"
    )

