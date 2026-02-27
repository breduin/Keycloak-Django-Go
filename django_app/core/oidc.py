from mozilla_django_oidc.auth import OIDCAuthenticationBackend


class KeycloakOIDCBackend(OIDCAuthenticationBackend):
    def create_user(self, claims):
        user = super().create_user(claims)
        user.keycloak_id = claims.get("sub")
        user.save()
        return user

    def update_user(self, user, claims):
        user = super().update_user(user, claims)
        user.keycloak_id = claims.get("sub")
        user.save()
        return user

