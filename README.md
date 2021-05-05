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
- rental
  - Index
    - PRIMARY             : (rental_id)
    - idx_fk_customer_id  : (customer_id)
    - idx_fk_inventory_id : (inventory_id)
    - rental_date         : (rental_date,inventory_id,customer_id)
    - idx_fk_staff_id     : (staff_id)
  - Cardinality
    - rental_id = 10000
    - inventory_id = 4575
    - customer_id = 599
    - return_date = 9800
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
- rental
  - Index
    - PRIMARY             : (rental_id)
    - idx_fk_customer_id  : (customer_id)
    - idx_fk_inventory_id : (inventory_id)
    - rental_date         : (rental_date,inventory_id,customer_id)
    - idx_fk_staff_id     : (staff_id)
  - Cardinality
    - return_date = 9800
    - rental_date = 9794
    - inventory_id = 4575
    - customer_id = 599

- customer
  - Index
    - PRIMARY             : (customer_id)
    - idx_fk_address_id   : (address_id)
    - idx_last_name       : (last_name)
    - idx_fk_store_id     : (store_id)
  - Cardinality
    - customer_id = 599
    - first_name = 591
    - last_name = 599
    - address_id = 599

- address
  - Index
    - PRIMARY             : (address_id)
    - idx_fk_city_id      : (city_id)
    - idx_location        : (location)
  - Cardinality
    - address_id = 603
    - phone = 602

- inventory
  - Index
    - PRIMARY             : (inventory_id)
    - idx_fk_film_id      : (film_id)
    - idx_store_id_film_id: (store_id,film_id)
  - Cardinality
    - inventory_id = 4581
    - film_id = 958

- film
  - Index
    - PRIMARY             : (film_id)
    - idx_fk_original_language_id: (original_language_id)
    - idx_title           : (title)
    - idx_fk_language_id  : (language_id)
  - Cardinality
    - rental_duration = 5
    - film_id = 1000
    - title = 1000
```

I refered and tested with [sakila sample database](https://dev.mysql.com/doc/sakila/en/) like above examples.


