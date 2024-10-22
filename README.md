# mdx - Execute your Markdown Code Blocks

Imagine you have the following Markdown file to document your commands:


    # demo.md
    ## [simple_echo]() - Simple echo in shell

    ```sh
    echo "hello world"
    ```

With `mdx` you execute the `sh` code block:

```
% mdx simple_echo
hello world
```

🚀 Find more examples [in the Wiki.](https://github.com/dim0x69/mdx/wiki/Examples)

## Getting started

### Installation

You can simply download a binary which fits your operating system and achitecture from the [releases page](https://github.com/dim0x69/mdx/releases).

Then just use the demo.md from above to execute `simple_echo`.

### Build

Go should be installed on your system: Follow [this guide](https://go.dev/doc/install) to install go in your PATH.

```sh
$ git clone https://github.com/dim0x69/mdx
$ go build
$ go install
```

## Resources
The idea for this project came from [Makedown](https://github.com/tzador/makedown).