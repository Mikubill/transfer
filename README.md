# Transfer
<a title="Release" target="_blank" href="https://github.com/Mikubill/transfer/releases"><img src="https://img.shields.io/github/release/Mikubill/transfer.svg?style=flat-square&hash=c7"></a>
<a title="Go Report Card" target="_blank" href="https://goreportcard.com/report/github.com/Mikubill/transfer"><img src="https://goreportcard.com/badge/github.com/Mikubill/transfer?style=flat-square"></a>

Simple Big File Transfer

🍭集合多个API的大文件传输工具

## install

Go语言程序, 可直接在[发布页](https://github.com/Mikubill/transfer/releases)下载使用。

或者使用安装脚本:

```shell
curl -sL https://git.io/file-transfer | sh 
```

## support

目前支持的服务:

|  Name   | Site  | Limit |
|  ----  | ----  |  ----  |
| Airportal | https://aitportal.cn/ | - |
| bitSend | https://bitsend.jp/ | - |
| CatBox | https://catbox.moe/ | 100MB |
| CowTransfer | https://www.cowtransfer.com/ | 2GB |
| GoFile | https://gofile.io/ | - |
| TmpLink | https://tmp.link/ | login only |
| Vim-cn | https://img.vim-cn.com/ | 100MB |
| WenShuShu | https://www.wenshushu.cn/ | 5GB |
| WeTransfer | https://wetransfer.com/ | 2GB |
| FileLink | https://filelink.io/ | - |
| Transfer.sh | https://transfer.sh/ | - |

开发中的服务

|  Name   | Site  | Limit |
|  ----  | ----  |  ----  |
| Firefox Send | https://send.firefox.com/ | 1GB |

## usage 

```shell

Usage:

  ./transfer command <backend> [options] file(s)/url(s)

Available Commands:
  download    Download a url or urls
  help        Help about any command
  upload      Upload a file or dictionary

Backend Support:
  arp - AirPortal https://airportal.cn/
  cow - Cowtransfer https://www.cowtransfer.com/
  wss - Wenshushu https://www.wenshushu.cn/
  bit - BitSend https://www.bitsend.jp/
  tmp - TmpLink https://tmp.link/
  cat - CatBox https://catbox.moe/
  vim - Vim-CN https://img.vim-cn.com/
  gof - GoFile https://gofile.io/
  wet - WeTransfer https://wetransfer.com/

Global Options:
  -v, --verbose               Verbose Mode
  -k, --keep                  Keep program active when upload/download finish

```

所有上传操作都建议指定一个API，如不指定将使用默认(filelink.Backend)。加上想要传输的文件/文件夹即可。

```shell
# upload
./transfer upload balabala.mp4

# upload
./transfer upload wss balabala.mp4

# upload folder
./transfer upload wet /path/
```

下载操作会自动识别支持的链接，不需要指定服务名称。

```
# download file
./transfer download https://.../
```

选定API以后不加链接或者文件，将显示关于该服务的相关信息：

```shell

➜ ./transfer upload cow
cowTransfer - https://cowtransfer.com/

  Size Limit:             2G(Anonymous), ~100G(Login)
  Upload Service:         qiniu object storage, East China
  Download Service:       qiniu cdn, Global

Usage:
  transfer upload cow [flags]

Aliases:
  cow, cow, cowtransfer

Flags:
      --block int         Upload block size (default 262144)
  -c, --cookie string     Your user cookie (optional)
      --hash              Check hash after block upload
  -h, --help              help for cow
  -p, --parallel int      Set the number of upload threads (default 4)
      --password string   Set password
  -s, --single            Upload multi files in a single link
  -t, --timeout int       Request retry/timeout limit in second (default 10)
      --verbose           Verbose mode to debug

Global Flags:
  -k, --keep      Keep program active when process finish
      --version   Show version and exit
```