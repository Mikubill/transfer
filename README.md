# Transfer
<a title="Release" target="_blank" href="https://github.com/Mikubill/transfer/releases"><img src="https://img.shields.io/github/release/Mikubill/transfer.svg?style=flat-square&hash=c7"></a>
<a title="Go Report Card" target="_blank" href="https://goreportcard.com/report/github.com/Mikubill/transfer"><img src="https://goreportcard.com/badge/github.com/Mikubill/transfer?style=flat-square"></a>

🍭集合多个API的大文件传输工具

Large file transfer tool with multiple file transfer services support

## install

Go语言程序, 可直接在[发布页](https://github.com/Mikubill/transfer/releases)下载使用。

或者使用安装脚本:

```shell
curl -sL https://git.io/file-transfer | sh 
```

Github Action中有实时构建版本，如有需要可以在Github Action的构建日志中查看下载链接。

## support

目前支持的文件传输服务:

|  Name   | Site  | Limit | Provider |
|  ----  | ----  |  ----  |  ----  |
| Airportal | https://aitportal.cn/ | - | Aliyun |
| bitSend | https://bitsend.jp/ | - | OVH |
| CatBox | https://catbox.moe/ | 100MB | Psychz |
| CowTransfer | https://www.cowtransfer.com/ | 2GB | Qiniu |
| GoFile | https://gofile.io/ | - | - |
| TmpLink | https://tmp.link/ | - | - |
| Vim-cn | https://img.vim-cn.com/ | 100MB | CloudFlare |
| WenShuShu | https://www.wenshushu.cn/ | 5GB | QCloud |
| WeTransfer | https://wetransfer.com/ | 2GB | CloudFront |
| FileLink | https://filelink.io/ | - | GCE |
| Transfer.sh | https://transfer.sh/ | - | Hetzner |
| Lanzous | https://www.lanzous.com/ | beta | - |

目前支持的图床服务:

|  Name   | Limit  | 
|  ----  | ----  |
| Ali | 5MB |
| Baidu | 10MB |
| CCUpload | 20MB (region limit) |
| Juejin | 20MB |
| Netease | 10MB |
| Prntscr | 10MB |
| SMMS | 5MB |
| Sugou | 10MB |
| Toutiao | - |
| Xiaomi | - |
| Suning | - |

开发中的服务

|  Name   | Site  | Limit |
|  ----  | ----  |  ----  |
| Firefox Send | https://send.firefox.com/ | 1GB |

## usage 

```shell

Transfer is a very simple big file transfer tool.

Backend Support:
  arp  -  Airportal  -  https://aitportal.cn/
  bit  -  bitSend  -  https://bitsend.jp/
  cat  -  CatBox  -  https://catbox.moe/
  cow  -  CowTransfer  -  https://www.cowtransfer.com/
  gof  -  GoFile  -  https://gofile.io/
  tmp  -  TmpLink  -  https://tmp.link/
  vim  -  Vim-cn  -  https://img.vim-cn.com/
  wss  -  WenShuShu  -  https://www.wenshushu.cn/
  wet  -  WeTransfer  -  https://wetransfer.com/
  flk  -  FileLink  -  https://filelink.io/
  trs  -  Transfer.sh  -  https://transfer.sh/

Usage:
  transfer [command]

Examples:
  # upload via wenshushu
  ./transfer wss <your-file>

  # download link
  ./transfer https://.../

Available Commands:
  decrypt     Decrypt a file
  encrypt     Encrypt a file
  hash        Hash a file
  help        Help about any command
  image       Upload a image to imageBed

Flags:
      --encrypt              encrypt stream when upload
      --encrypt-key string   specify the encrypt key
  -f, --force                attempt to download file regardless error
  -h, --help                 help for transfer
      --keep                 keep program active when process finish
  -o, --output string        download to another file/folder (default ".")
  -p, --parallel int         set download task count (default 3)
  -t, --ticket string        set download ticket
  -v, --verbose              enable verbose mode to debug
      --version              show version and exit

Use "transfer [command] --help" for more information about a command.
```

### upload & download

所有上传操作都建议指定一个API，如不指定将使用默认(filelink.Backend)。加上想要传输的文件/文件夹即可。

```shell 

Upload a file or folder.

Usage:
  transfer [flags] <files>

Aliases:
  upload, up

Available APIs:
  arp  -  Airportal  -  https://aitportal.cn/
  bit  -  bitSend  -  https://bitsend.jp/
  cat  -  CatBox  -  https://catbox.moe/    
  cow  -  CowTransfer  -  https://www.cowtransfer.com/  
  gof  -  GoFile  -  https://gofile.io/
  tmp  -  TmpLink  -  https://tmp.link/     
  vim  -  Vim-cn  -  https://img.vim-cn.com/    
  wss  -  WenShuShu  -  https://www.wenshushu.cn/  
  wet  -  WeTransfer  -  https://wetransfer.com/  
  flk  -  FileLink  -  https://filelink.io/
  trs  -  Transfer.sh  -  https://transfer.sh/

Flags:
      --encrypt              Encrypt stream when upload
      --encrypt-key string   Specify the encrypt key
  -h, --help                 help for upload

Global Flags:
      --keep      Keep program active when process finish
      --version   Show version and exit

Use "transfer upload [command] --help" for more information about a command.
```

Examples

```shell
# upload
./transfer balabala.mp4

# upload
./transfer wss balabala.mp4

# upload folder
./transfer wet /path/
```

不同的Backend提供不同的选项，可以在帮助中查看关于该服务的相关信息。

```shell 
➜  ./transfer cow

cowTransfer - https://cowtransfer.com/

  Size Limit:             2G(Anonymous), ~100G(Login)
  Upload Service:         qiniu object storage, East China
  Download Service:       qiniu cdn, Global

Usage:
  transfer cow [flags]

Aliases:
  cow, cow, cowtransfer

Flags:
      --block int         Upload block size (default 262144)
  -c, --cookie string     Your user cookie (optional)
      --hash              Check hash after block upload
  -h, --help              help for cow
  -p, --parallel int      Set the number of upload threads (default 2)
      --password string   Set password
  -s, --single            Upload multi files in a single link
  -t, --timeout int       Request retry/timeout limit in second (default 10)

Global Flags:
      --encrypt              encrypt stream when upload
      --encrypt-key string   specify the encrypt key
      --keep                 keep program active when process finish
  -v, --verbose              enable verbose mode to debug
      --version              show version and exit
```

下载操作会自动识别支持的链接，不需要指定服务名称。

```
# download file
./transfer download https://.../
```

试验性功能：`--encrypt`选项可以在上传时将文件加密，下载时需要配合`--decrypt`选项才能正确下载文件。（当然也可以先下载后再解密）加密方式为AES-CBC，默认会自动生成一个密码，也可以通过`--encrypt-key`指定一个。

```shell 
# encrypt stream when upload
➜ ./transfer wss --encrypt transfer
Warning: crypto mode is enabled.
Note: Crypto mode still in beta and abnormalities may occur, do not over-rely on this function.
Key is not set or incorrect: Setting it to 94d0500605b372245dc77f95fbc20010
...

# encrypt with key
➜ ./transfer wss --encrypt --encrypt-key=123 transfer
Warning: crypto mode is enabled.
Note: Crypto mode still in beta and abnormalities may occur, do not over-rely on this function.
Encrypt using key: 123
...

# decrypt stream when download
➜ ./transfer --encrypt --encrypt-key=123 https://....
Warning: crypto mode is enabled.
Note: Crypto mode is not compatible with multi thread download mode, setting parallel to 1.
...
```

### image

transfer也支持上传图片至图床，默认自动使用阿里图床上传，也可以通过`-b, --backend`指定图床。

```shell 

Upload a image to imageBed.
Default backend is ali.backend, you can modify it by -b flag.

Backend support:
  alibaba(ali), baidu(bd), ccupload(cc), juejin(jj),
  netease(nt), prntscr(pr), smms(sm), sogou(sg),
  toutiao(tt), xiaomi(xm), vimcn(vm), suning(sn)

Example:
  # simply upload
  transfer image your-image

  # specify backend to upload
  transfer image -b sn your-image

Note: Image bed backend may have strict size or format limit.

Usage:
  transfer image [flags]

Flags:
  -b, --backend string   Set upload/download backend
  -h, --help             help for image

Global Flags:
      --encrypt              encrypt stream when upload
      --encrypt-key string   specify the encrypt key
      --keep                 keep program active when process finish
  -v, --verbose              enable verbose mode to debug
      --version              show version and exit
```

### encrypt & decrypt

和前面upload使用的是同样的加密，只是在本地进行。也可以使用前面下载的加密后文件在此解密。可以通过不同参数指定密钥和输出文件名

关于加密的说明：目前只能选择AES-CBC的加密方式，分块大小策略为min(1m, fileSize)

```shell 
# encrypt
transfer encrypt your-file

# encrypt using specified key
transfer encrypt -k abc your-file

# decrypt using specified key
transfer decrypt -k abc your-file

# specify path
transfer encrypt -o output your-file
```

### hash 

hash功能使用sha1, crc32, md5, sha256对文件进行校验，可以用来检验文件一致性。

```shell 
➜  ./transfer hash main.go
size: 68
path: /../transfer/main.go

crc32: a51da8f5
md5: aa091bb918ab85b1dc44cb771b1663d1
sha1: a8e25d41330c545da8bcbeade9aebdb1b4a13ab7
sha256: ab4dd3cdd79b5e2a88fcb3fcd45dfcffc935c913adfa888f3fb50b324638e958
```
