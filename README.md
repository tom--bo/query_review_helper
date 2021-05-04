# Query Review Helper


- クエリとdb情報をなげるだけ.
- table alias nameが合ってもoriginalのtable nameがとれる
- クエリに含まれるテーブル一覧、そのindex一覧が取れる

## How to use

```
$ go build

$ ./query_review_helper -u sysbench -p sysbench -h 192.168.1.50 -P 3306  -d sample
(Input query and Ctrl-D at the last line)
Select * from
t2 where c1 = 1
(Ctrl-D)
```
