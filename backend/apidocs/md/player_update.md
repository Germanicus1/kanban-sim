## UpdatePlayer handles HTTP PUT requests to update a player's information

It validates the request method, decodes the request payload, and ensures the payload contains valid player data. If the data is valid, it calls the service layer to update the player in the database. In case of errors, it responds with appropriate HTTP status codes and error messages.
