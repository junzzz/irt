image resize tool
==================

[nfnt/resize](https://github.com/nfnt/resize)を使ってCUIで画像リサイズできる    
透過画像未対応


### インストール

```
$ go get github.com/junzzz/irt/cmd/irt
```

### 使い方

```
$ irt -w 50% -o /path/to/resize_image.jpg /path/to/image.jpg
$ irt -h 100px /path/to/image.gif
$ irt -l 30% /path/to/directory/
```

