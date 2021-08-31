# migrator performance

Back in 2018 I wrote a small benchmark to compare 3 DB migrations frameworks:

- migrator
- proprietary Ruby framework - used internally at my company
- flyway - leading market feature rich DB migration framework: https://flywaydb.org

In 2021 I decided to refresh the results (the original results for historical purposes are at the bottom).

You can play around with performance benchmark yourself. See `test/performance/test.sh` for migrator test and `test/performance/flywaydb-test.sh` for flyway.

# 2021 results

In 2021 I compared migrator (`v2021.1.0`) and flyway (`Flyway Teams Edition 7.14.0 by Redgate`).

_Note: I didn't use Ruby framework as it became deprecated. Also, the previous Ruby's results were dramatic and not worth investing time in re-doing them._

Compared to 2018 the tests are single tenant with 10k migrations. This is because somewhere between 2018 and 2021 [multiple schemas](https://flywaydb.org/documentation/learnmore/faq.html#multiple-schemas) stopped working in flyway.

| framework | number of migrations | time (s) | memory consumption (MB) |
| --------- | -------------------- | -------- | ----------------------- |
| migrator  | 10000                | 5        | 23                      |
| flyway    | 10000                | 49       | 265                     |

migrator is still orders of magnitude better than flyway.

Because multiple schemas stopped working in flyway there wasn't too much sense in doing more comparison benchmarks.

Instead, I ran additional multi-tenant/multiple-schema migrator benchmarks.

# migrator performance showcase

You can use `test/performance/test.sh` to run any simulation you want. You can also use it to simulate adding new migrations (so called append mode) - scroll to the bottom of that script to see a comment showing you how to do this.

I prepared 2 multi-tenant simulations:

1. 1000 tenants, 10 versions, 20 SQL files in each version - 20k migrations to apply in each version
2. 500 tenants, 5 versions, 100 SQL files in each version - 50k migrations to apply in each version

## 1000 tenants

Execution time is growing slightly with every new version. The memory consumption grows proportionally to how many migrations are in the database. This is because migrator fetches all migrations from database to compute which migrations were already applied and which are to be applied.

| version | number of migrations (before - after) | time (s) | memory consumption (MB) |
| ------- | ------------------------------------- | -------- | ----------------------- |
| 1       | 0 - 21001                             | 57       | 66                      |
| 2       | 21001 - 41001                         | 58       | 86                      |
| 3       | 41001 - 61001                         | 56       | 101                     |
| 4       | 61001 - 81001                         | 62       | 165                     |
| 5       | 81001 - 101001                        | 62       | 175                     |
| 6       | 101001 - 121001                       | 59       | 242                     |
| 7       | 121001 - 141001                       | 71       | 280                     |
| 8       | 141001 - 161001                       | 68       | 300                     |
| 9       | 161001 - 181001                       | 70       | 324                     |
| 10      | 181001 - 201001                       | 69       | 380                     |

## 500 tenants

Similarly to 1000 tenants, 500 tenants simulation execution time is growing slightly with every new version. The memory consumption grows proportionally to how many migrations are in the database.

Based on both simulations we can see that migrator under any load is stable and behaves very predictable:

- applied ~300 migrations a second (creating tables, inserting multiple rows, database running on the same machine = my MacBook)
- consumed ~2.5MB memory for every 1k migrations

| version | number of migrations (before - after) | time (s) | memory consumption (MB) |
| ------- | ------------------------------------- | -------- | ----------------------- |
| 1       | 0 - 50501                             | 167      | 126                     |
| 2       | 50501 - 100501                        | 170      | 140                     |
| 3       | 100501 - 150501                       | 167      | 218                     |
| 4       | 150501 - 200501                       | 181      | 292                     |
| 5       | 200501 - 250501                       | 178      | 396                     |

# 2018

Keeping 2018 results for historical purposes.

_Note: For all 3 DB frameworks I used multiple schemas benchmark. Unfortunately, back in 2018, I didn't commit flyway tests and its impossible now to recreate how multiple schemas were actually set up._

Execution times were the following:

| # Tenants | # Existing Migrations | # Migrations to apply | migrator | Ruby | Flyway |
| --------- | --------------------- | --------------------- | -------- | ---- | ------ |
| 10        | 0                     | 10001                 | 154s     | 670s | 2360s  |
| 10        | 10001                 | 20                    | 2s       | 455s | 340s   |

migrator was the undisputed winner.

The Ruby framework had the undesired functionality of making a DB call each time to check if given migration was already applied. migrator fetched all applied migrations at once and compared them in memory. That was the primary reason why migrator was so much better in the second test.

flyway results were... very surprising. I was so shocked that I had to re-run flyway as well as all other tests. Yes, flyway was 15 times slower than migrator in the first test. In the second test flyway was faster than Ruby. Still a couple orders of magnitude slower than migrator.

The other thing to consider is the fact that migrator is written in go which is known to be much faster than Ruby and Java.
