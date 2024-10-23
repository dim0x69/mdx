## [cmd1](dep1 dep2)

This test would output Hello, if the availability of all deps is not validated before execution.

```sh
echo -n "Hello"
```
## [dep1]()

```sh
echo -n "Hello"
```

## [dep2](dep2.1)

```sh
echo -n "!"
```