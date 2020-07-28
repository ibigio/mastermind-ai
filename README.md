# Mastermind AI
An ever-so-slightly superhuman Mastermind player.

Each guess maximizes the expected number of potential secret codes to discard. Fully optimized to run in parallel on however many CPUs are available. 

**Rules to Mastermind:** [Wikipedia](https://en.wikipedia.org/wiki/Mastermind_(board_game)#:~:text=Each%20guess%20is%20made%20by,in%20both%20color%20and%20position.)

## Build and Run

```bash
go build
./mastermind-ai
```

## Performance

To run a performance evaluation on the AI, simply call the `runEvaluation` function. Recommended number of games is 20 - 100.