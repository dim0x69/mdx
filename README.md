# mdx - Execute your Markdown Code Blocks

Imagine you have the following Markdown  file to document your commands:


    # demo.md
    ## `simple_echo` - Simple echo in shell

    ```sh
    echo "{{.arg1}} {{.arg2}}"
    ```

With `mdx` you execute the `sh` code block:

```
% ./mdx simple_echo hello world
hello world
```

## Resources
The idea for this project came from [Makedown](https://github.com/tzador/makedown).