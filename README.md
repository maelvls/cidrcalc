# cidrcalc

CLI for calculating the biggest CIDR from a list of IPs.

Install:

```bash
go install github.com/maelvls/cidrcalc@latest
```

Examples:

```console
$ dig +short clouddocsdev.s3-website-us-west-2.amazonaws.com | grep -Eo '([0-9]{1,3}\.){3}[0-9]{1,3}' | cidrcalc
Largest CIDR block: 52.0.0.0/8
```

Alternatively, use `-hostname` flag to resolve the IPs from a hostname without
having to use `dig`:

```console
$ cidrcalc -hostname clouddocsdev.s3-website-us-west-2.amazonaws.com
Resolved IPs for clouddocsdev.s3-website-us-west-2.amazonaws.com: [52.218.153.66 52.218.252.146 52.92.225.227 52.218.188.3 52.92.206.163 52.92.204.171 52.92.213.51 52.92.209.3]
Largest CIDR block: 52.0.0.0/8
```
