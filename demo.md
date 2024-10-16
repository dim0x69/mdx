# demo.md
## `simple_echo` - Simple echo in shell

```sh
echo "{{.arg1}} {{.arg2}}"
```

Execute with `mdx demo.md simple_echo hello world`

Output:
```
$ hello world
```




## `shebang1`- Example using shebang from a python venv

```
#!/home/ldm9fe/git/mdx/.venv/bin/python

import sys
print(sys.executable)
```

Note: No infostring is specified.