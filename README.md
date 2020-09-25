# Adventure Land PoC

A proof-of-concept bot for [Adventure Land][adventureland].

## Preface

This bot doesn't do anything. As in, there is no AI logic.
This is (intentionally) a skeleton project.

Programming the AI is **is how you play Adventure Land**,
and my goal is not to deprive anyone of that joy.

## Then what is this?

This skeleton is a good starting point for anyone who wants
to write an AI for Adventure Land using Go.

This project uses the [github.com/mxschmitt/playwright-go][playwright-go]
library to interface with [Playwright][playwright], and ultimately,
headless chromium.

Running the JS runtime in a headless browser has advantages,
namely performance.

...But I hate JavaScript, and even TypeScript is a pain. :)

## How do I build this?

```sh
$ go run main.go
```

[adventureland]: https://adventure.land/
[playwright-go]: https://github.com/mxschmitt/playwright-go
[playwright]: https://github.com/microsoft/playwright
