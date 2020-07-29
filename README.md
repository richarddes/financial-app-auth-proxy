# Auth-Proxy
This service handles all services that don't need an authenticated user (login, register, etc.). It also provides a reverse-proxy which first checks if the user's authenticated and then redirects to the specific server.

## Setup 
You can run this app as a standalone app even though you need to have a DNS that resolves *news_service*, *stock_service* and *user_service*, since those values are hard-coded. 
To install all necessary dependencies, run:
```sh
go get ./...
```
To run the app, run:
```sh
go run main.go
```

## Inner workings
All POST request to the service first go through a csrf middleware. Afterwards all requests staring with */api/users*, */api/news* or */api/stocks* go through an authentication middleware, that checks that the user has a valid authentication token. Once they passed the middleware, those request are being redirected to their specific service.

The following 4 endpoints are the only ones' that are directly handled by the *Auth-Proxy*
- */register* handles registrations
- */login* handles logins
- */get-csrf-token* returns a new csrf token 
- */check-credentials* checks if the user already has valid credentials. If so it send a status code 200 (Ok). It uses the authentication midleware under the hood.

## Contributing 
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
MIT License. Click [here](https://choosealicense.com/licenses/mit/) or see the LICENSE file for details.