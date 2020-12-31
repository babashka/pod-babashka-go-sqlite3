#!/usr/bin/env bb

(ns honeysql
  (:require [babashka.deps :as deps]
            [babashka.pods :as pods]
            [clojure.java.io :as io]))

(deps/add-deps '{:deps {honeysql/honeysql {:mvn/version "1.0.444"}}})

(require '[honeysql.core :as sql]
         '[honeysql.helpers :as helpers])

(pods/load-pod "./pod-babashka-sqlite3")

(require '[pod.babashka.sqlite3 :as sqlite])

(.delete (io/file "/tmp/foo.db"))

(prn (sqlite/execute! "/tmp/foo.db" ["create table if not exists foo (col1 TEXT, col2 TEXT)"]))
(prn (sqlite/execute! "/tmp/foo.db" ["delete from foo"]))

(def insert
  (-> (helpers/insert-into :foo)
      (helpers/columns :col1 :col2)
      (helpers/values
       [["Foo" "Bar"]
        ["Baz" "Quux"]])
      sql/format))

(prn insert)

(prn (sqlite/execute! "/tmp/foo.db" insert))

(def sqlmap {:select [:col1 :col2]
             :from   [:foo]
             :where  [:= :col1 "Foo"]})

(def sql (sql/format sqlmap))

(prn sql)

(def results (sqlite/query "/tmp/foo.db" sql))
(prn results)

