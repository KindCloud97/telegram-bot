# Telegram-Bot

## About the project

This project consists of two parts:

- Telegram bot that processes text messages and images.
- GRPC server API for the ability to connect to the bot operator.

Currently, telegram bot and the API are monolith, but its possible to separate them.

## Getting Started

To start bot and API run `go run cmd/server`. This will start telegram bot in long polling mode and grpc server on port `:8080`.

For testing purposes run `go run cmd/client`. The client can send text messeges and receive text or image messages from users.
> NOTE: The client uses kitty [terminal graphics protocol](https://sw.kovidgoyal.net/kitty/graphics-protocol/) to display images inside a terminal, that is, to see images you should use supported terminal.
