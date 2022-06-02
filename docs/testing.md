This document helps you to do some tests about the KubeSphere DevOps.

## Pressure test

Before get stated, we need to install some tools. In this case, I use [hd](https://github.com/LinuxSuRen/http-downloader/) to install them.

```shell
hd fetch
hd i kd
hd i names
```

The following command be able to create a lot of Pipelines:
```shell
for a in {1..1000}
do
  ks pip create --ws simple --project test --template simple --name $(names)
done
```
then, run all the Pipelines

```shell

for a in $(ks get pipeline -n testkjhx9 -o custom-columns=Name:metadata.name)
do
  ks pip run  -n testkjhx9 -p  $a -b
done
```
