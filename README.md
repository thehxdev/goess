# Goess
Simple number guessing game through TCP.
Just to practice and learn network programming.

## Run
To run the server:
```bash
go run .
```

## Connect to server
To connect as clinet use netcat (`nc` command):
```bash
nc 127.0.0.1 8000
```

Then server responds with:
```
Welcome to goess game!
Guess a number between 1 and 100 (Both ends are included)
You only have 10 guesses!

Guess 1 =>
```

Happy gaming :)
