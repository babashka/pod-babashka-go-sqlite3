# pod-babashka-sqlite3

A [babashka pod](https://github.com/babashka/babashka.pods) for interacting with [sqlite3](https://www.sqlite.org/index.html).

Implemented using the Go [go-sqlite3](https://github.com/mattn/go-sqlite3) library.

## Status

Experimental.

## Usage

Load the pod and `pod.babashka.sqlite3` namespace:

``` clojure
(ns sqlite3-script
  (:require [babashka.deps :as deps]))

(pods/load-pod 'org.babashka/sqlite3 "0.0.1")
(require '[pod.babashka.sqlite3 :as sqlite])
```

The namespace exposes two functions: `execute!` and `query`. Both accept a path
to the sqlite database and a query vector:

``` clojure
(sqlite/execute! "/tmp/foo.db"
  ["create table if not exists foo (the_text TEXT, the_int INTEGER, the_real REAL, the_blob BLOB)"])

;; This pod also supports storing blobs, so lets store a picture.
(def png (java.nio.file.Files/readAllBytes (.toPath (io/file "resources/babashka.png"))))

(sqlite/execute! "/tmp/foo.db"
  ["insert into foo (the_text, the_int, the_real, the_blob) values (?,?,?,?)" "foo" 1 3.14 png])
;;=> {:rows-affected 1, :last-inserted-id 1}

(def results (sqlite/query "/tmp/foo.db" ["select * from foo order by the_int asc"]))
(count results) ;;=> 1

(def row (first results))
(keys row) ;;=> (:the_text :the_int :the_real :the_blob)
(:the_text row) ;;=> "foo"

;; Should be true:
(= (count png) (count (:the_blob row)))
```

See [test/script.clj](test/script.clj) for an example test script.

### HoneySQL

``` clojure
(ns honeysql-script
  (:require [babashka.deps :as deps]
            [babashka.pods :as pods]))

;; Load HoneySQL from Clojars:
(deps/add-deps '{:deps {honeysql/honeysql {:mvn/version "1.0.444"}}})

(require '[honeysql.core :as sql]
         '[honeysql.helpers :as helpers])

(pods/load-pod 'org.babashka/sqlite3 "0.0.1")
(require '[pod.babashka.sqlite3 :as sqlite])

(sqlite/execute! "/tmp/foo.db" ["create table if not exists foo (col1 TEXT, col2 TEXT)"])

(def insert
  (-> (helpers/insert-into :foo)
      (helpers/columns :col1 :col2)
      (helpers/values
       [["Foo" "Bar"]
        ["Baz" "Quux"]])
      sql/format))
;; => ["INSERT INTO foo (col1, col2) VALUES (?, ?), (?, ?)" "Foo" "Bar" "Baz" "Quux"]

(sqlite/execute! "/tmp/foo.db" insert)
;; => {:rows-affected 2, :last-inserted-id 2}

(def sqlmap {:select [:col1 :col2]
             :from   [:foo]
             :where  [:= :col1 "Foo"]})

(def select (sql/format sqlmap))
;; => ["SELECT col1, col2 FROM foo WHERE col1 = ?" "Foo"]

(sqlite/query "/tmp/foo.db" select)
;; => [{:col1 "Foo", :col2 "Bar"}]
```

See [test/honeysql.clj](test/honeysql.clj) for a HoneySQL example script.

## Build

### Requirements

- [Go](https://golang.org/dl/) 1.15+ should be installed.
- Clone this repo.
- Run `go build -o pod-babashka-sqlite3 main.go` to compile the binary.

## License

Copyright © 2020 Michiel Borkent and Rahul De

License: [BSD 3-Clause](https://opensource.org/licenses/BSD-3-Clause)
