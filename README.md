# Transfer
<a title="Release" target="_blank" href="https://github.com/Mikubill/transfer/releases"><img src="https://img.shields.io/github/release/Mikubill/transfer.svg?style=flat-square&hash=c7"></a>
<a title="Go Report Card" target="_blank" href="https://goreportcard.com/report/github.com/Mikubill/transfer"><img src="https://goreportcard.com/badge/github.com/Mikubill/transfer?style=flat-square"></a>

Simple Big File Transfer

🍭集合多个API的大文件传输工具

## install

Go语言程序, 可直接在[发布页]((https://github.com/Mikubill/cowtransfer-uploader/releases))下载使用。

或者使用安装脚本:

```shell
curl -sL https://git.io/file-transfer | sh 
```

## support

目前支持的服务:

|  Name   | Site  | Limit |
|  ----  | ----  |  ----  |
| bitSend | https://bitsend.jp/ | - |
| CatBox | https://catbox.moe/ | 100MB |
| CowTransfer | https://www.cowtransfer.com/ | 2GB |
| GoFile | https://gofile.io/ | - |
| TmpLink | https://tmp.link/ | login only |
| Vim-cn | https://img.vim-cn.com/ | 100MB |
| WenShuShu | https://www.wenshushu.cn/ | 5GB |
| WeTransfer | https://wetransfer.com/ | 2GB |

开发中的服务

|  Name   | Site  | Limit |
|  ----  | ----  |  ----  |
| Firefox Send | https://send.firefox.com/ | 1GB |

## usage 

```shell

Usage:

  ./transfer <backend> [options] file(s)/url(s)

Backend Support:

  cow - Cowtransfer https://www.cowtransfer.com/
  wss - Wenshushu https://www.wenshushu.cn/
  bit - BitSend https://www.bitsend.jp/
  tmp - TmpLink https://tmp.link/
  cat - CatBox https://catbox.moe/
  vim - Vim-CN https://img.vim-cn.com/
  gof - GoFile https://gofile.io/
  wet - WeTransfer https://wetransfer.com/

Global Options:

  -k, --keep                  Keep program active when upload/download finish

```

所有操作都需要指定一个服务。指定好上传服务后加上想要传输的文件/文件夹/链接即可。

```shell
# upload
./transfer balabala.mp4

# upload folder
./transfer /path/

# download file
./transfer https://.../
```
