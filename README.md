
# Redis Migration tool

this util supposed to migrate list keys type from one redis instance to another redis instance 

some cloud providers have a limitation for migration as redis migration commands are not available

I need this specifically for lists but it can work with anything with minor changes.

to execute the migration please follow the following example commandline

```shell
go run main.go -src="src-host:src-port" -dst="dst-host:dst-port" -key="*key-pattern*" -ttl=3600
```