
Authproxy is part of middleend and it exposes following 3 apis
1. **/v1/login**
   - Redirects user to keycloak login page.
   - Sets a cookie with original URL
2. **/v1/callback**
   - After successful login gets auth code and exchange it for token.
   - Set id_token and access_token in cookie and redirects to original URL
3. **/v1/auth**
   - Retrieve idtoken from cookie and verifies the JWT.
   - If id_token is valid then access to resources else redirects to login page.

Required inputs of authproxy comes from authproxy section of helm config
- Issuer
- Redirect URI
- Client id
