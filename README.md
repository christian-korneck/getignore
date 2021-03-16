# get gitignore templates in the shell

`getignore` is a CLI client for [GitHub's .gitignore templates](https://github.com/github/gitignore). List and print gitignore templates for a wide variety of languages from the terminal.

List available languages with:
```
getignore --list
```

and easily bootstrap or extend a `.gitingore` file:

```
getignore python go visualstudiocode macos >> .gitignore
```

### Motivation

There are similar tools [[1]](https://github.com/aswinkarthik/gitignore.cli) [[2]](https://github.com/vccolombo/download-gitignore) but they require a script interpreter. I wanted a static executable that I can copy to arbitrary [vscode devcontainers](https://code.visualstudio.com/docs/remote/containers). 



