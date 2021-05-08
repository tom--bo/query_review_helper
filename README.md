# Query Review Helper

This tool is still in beta version.  
Only support MySQL compatible dataabses.

Get table's indexes and estimated column cardinality from query to review.  
Please use this with `EXPLAIN` check.

column cardinality is estimated by first and last 5000 rows in PK.
Like ...
```sql
SELECT count(distinct {column})
FROM (
	(SELECT {column} FROM {table} ORDER BY {PK_COL1} ASC, {PK_COL2} ASC LIMIT 5000)
	UNION DISTINCT
	(SELECT {column} FROM {table} ORDER BY {PK_COL1} DESC, {PK_COL2} DESC LIMIT 5000)
) tmp
```

These queries and implementations will be improved! (again, this is still in beta)


## How to use

input query from stdin and ^D at last.  
It shows `EXPLAIN` result and index/cardinality information referring your query.  
You can omit output with some options (`-e`, `-i`, `-c`)

```sh
$ go build -o bin/query_review_helper

$ ./bin/query_review_helper -u {user} -p {password} -h 192.168.1.10 -P 3306  -d sakila
(Input query and Ctrl-D at the last line)
SELECT rental_id
           FROM rental
           WHERE inventory_id = 10
           AND customer_id = 3
           AND return_date IS NULL
^D

==== Explain Result ====
+----+-------------+----------------------+--------+-------------+----------------------------------+--------+------------------------------------------+----------+----------+------------+
| id | select_type |     table            | part   |   type      |         key                      | keylen |    ref                                   |   rows   | filtered |  extra     |
+----+-------------+----------------------+--------+-------------+----------------------------------+--------+------------------------------------------+----------+----------+------------+
|  1 | SIMPLE      | rental               | NULL   | index_merge | idx_fk_inventory_id,idx_fk_customer_id | 3,2    | NULL                                     | 1        |   10.00 | Using intersect(idx_fk_inventory_id,idx_fk_customer_id); Using where |
+----+-------------+----------------------+--------+-------------+----------------------------------+--------+------------------------------------------+----------+----------+------------+


==== Tables in query ====

- rental
  - Index
    - PRIMARY             : (rental_id)
    - idx_fk_customer_id  : (customer_id)
    - idx_fk_inventory_id : (inventory_id)
    - rental_date         : (rental_date,inventory_id,customer_id)
    - idx_fk_staff_id     : (staff_id)
  - Cardinality
    - rental_id    = 10000
    - inventory_id = 4575
    - customer_id  = 599
    - return_date  = 9800
```


This also support aliases

```sh
$ ./bin/query_review_helper -u {user} -p {password} -h 192.168.1.10 -P 3306  -d sakila
(Input query and ^D at the last line)

SELECT CONCAT(c.last_name, ', ', c.first_name) AS customer,
           a.phone, f.title
           FROM rental r INNER JOIN customer c ON r.customer_id = c.customer_id
           INNER JOIN address a ON c.address_id = a.address_id
           INNER JOIN inventory i ON r.inventory_id = i.inventory_id
           INNER JOIN film f ON i.film_id = f.film_id
           WHERE r.return_date IS NULL
           AND rental_date + INTERVAL f.rental_duration DAY < CURRENT_DATE()
           ORDER BY title
           LIMIT 5;
^D

==== Explain Result ====
+----+-------------+----------------------+--------+-------------+----------------------------------+--------+------------------------------------------+----------+----------+------------+
| id | select_type |     table            | part   |   type      |         key                      | keylen |    ref                                   |   rows   | filtered |  extra     |
+----+-------------+----------------------+--------+-------------+----------------------------------+--------+------------------------------------------+----------+----------+------------+
|  1 | SIMPLE      | r                    | NULL   | ALL         | NULL                             | NULL   | NULL                                     | 16008    |   10.00 | Using where; Using temporary; Using filesort |
|  1 | SIMPLE      | c                    | NULL   | eq_ref      | PRIMARY                          | 2      | sakila.r.customer_id                     | 1        |   100.00 | NULL       |
|  1 | SIMPLE      | a                    | NULL   | eq_ref      | PRIMARY                          | 2      | sakila.c.address_id                      | 1        |   100.00 | NULL       |
|  1 | SIMPLE      | i                    | NULL   | eq_ref      | PRIMARY                          | 3      | sakila.r.inventory_id                    | 1        |   100.00 | NULL       |
|  1 | SIMPLE      | f                    | NULL   | eq_ref      | PRIMARY                          | 2      | sakila.i.film_id                         | 1        |   100.00 | Using where |
+----+-------------+----------------------+--------+-------------+----------------------------------+--------+------------------------------------------+----------+----------+------------+


==== Tables in query ====

- rental
  - Index
    - PRIMARY             : (rental_id)
    - idx_fk_customer_id  : (customer_id)
    - idx_fk_inventory_id : (inventory_id)
    - rental_date         : (rental_date,inventory_id,customer_id)
    - idx_fk_staff_id     : (staff_id)
  - Cardinality
    - rental_date  = 9794
    - inventory_id = 4575
    - customer_id  = 599
    - return_date  = 9800

- customer
  - Index
    - PRIMARY             : (customer_id)
    - idx_fk_address_id   : (address_id)
    - idx_last_name       : (last_name)
    - idx_fk_store_id     : (store_id)
  - Cardinality
    - customer_id  = 599
    - first_name   = 591
    - last_name    = 599
    - address_id   = 599

- address
  - Index
    - PRIMARY             : (address_id)
    - idx_fk_city_id      : (city_id)
    - idx_location        : (location)
  - Cardinality
    - phone        = 602
    - address_id   = 603

- inventory
  - Index
    - PRIMARY             : (inventory_id)
    - idx_fk_film_id      : (film_id)
    - idx_store_id_film_id: (store_id,film_id)
  - Cardinality
    - inventory_id = 4581
    - film_id      = 958

- film
  - Index
    - PRIMARY             : (film_id)
    - idx_fk_language_id  : (language_id)
    - idx_fk_original_language_id: (original_language_id)
    - idx_title           : (title)
  - Cardinality
    - film_id      = 1000
    - title        = 1000
    - rental_duration = 5
```

You can omit output with some options (`-e`, `-i`, `-c`).  
For example with `-e` option, you can only see `EXPLAIN` result.

```sql
$ ./bin/query_review_helper -u {user} -p {password} -h 192.168.1.10 -P 3306  -d sakila
(Input query and ^D at the last line)
SELECT CONCAT(c.last_name, ', ', c.first_name) AS customer,
       a.phone, f.title
FROM rental r INNER JOIN customer c ON r.customer_id = c.customer_id
              INNER JOIN address a ON c.address_id = a.address_id
              INNER JOIN inventory i ON r.inventory_id = i.inventory_id
              INNER JOIN film f ON i.film_id = f.film_id
WHERE r.return_date IS NULL
  AND rental_date + INTERVAL f.rental_duration DAY < CURRENT_DATE()
ORDER BY title
    LIMIT 5;

^D


==== Explain Result ====
+----+-------------+----------------------+--------+-------------+----------------------------------+--------+------------------------------------------+----------+----------+------------+
| id | select_type |     table            | part   |   type      |         key                      | keylen |    ref                                   |   rows   | filtered |  extra     |
+----+-------------+----------------------+--------+-------------+----------------------------------+--------+------------------------------------------+----------+----------+------------+
|  1 | SIMPLE      | r                    | NULL   | ALL         | NULL                             | NULL   | NULL                                     | 16008    |   10.00 | Using where; Using temporary; Using filesort |
|  1 | SIMPLE      | c                    | NULL   | eq_ref      | PRIMARY                          | 2      | sakila.r.customer_id                     | 1        |   100.00 | NULL       |
|  1 | SIMPLE      | a                    | NULL   | eq_ref      | PRIMARY                          | 2      | sakila.c.address_id                      | 1        |   100.00 | NULL       |
|  1 | SIMPLE      | i                    | NULL   | eq_ref      | PRIMARY                          | 3      | sakila.r.inventory_id                    | 1        |   100.00 | NULL       |
|  1 | SIMPLE      | f                    | NULL   | eq_ref      | PRIMARY                          | 2      | sakila.i.film_id                         | 1        |   100.00 | Using where |
+----+-------------+----------------------+--------+-------------+----------------------------------+--------+------------------------------------------+----------+----------+------------+


==== Tables in query ====
- rental
- customer
- address
- inventory
- film
```


## Options

```sh
Usage of ./query_review_helper:
  -P int
    	mysql port (default 3306)
  -S string
    	mysql unix domain socket
  -c	show cardinalities
  -d string
    	mysql database
  -debug
    	DEBUG mode
  -e	show explain results
  -f string
    	conf file for auth info
  -h string
    	mysql host (default "localhost")
  -i	show indexes
  -l int
    	limitation for cardinality sampling (default 5000)
  -p string
    	mysql password (default "password")
  -u string
    	mysql user (default "mysql")
```

### Config file

You can configure config-file with `-f` option.  
Currently, you can specify `user`, `password`, `port` in config file.

(Please see conf/sample.cnf)

```toml
[auth]
user     = "sample-user"
password = "passwd"
port     = 3306
```


### Note

I refered and tested with [sakila sample database](https://dev.mysql.com/doc/sakila/en/) like above examples.



