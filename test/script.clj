#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]
            [clojure.java.io :as io]
            [clojure.test :as t :refer [deftest is]]))

(prn (pods/load-pod "./pod-babashka-sqlite3"))

(require '[pod.babashka.sqlite3 :as sqlite])

(.delete (io/file "/tmp/foo.db"))

(prn (sqlite/execute! "/tmp/foo.db" ["create table if not exists foo (the_text TEXT, the_int INTEGER, the_real REAL)"]))
(prn (sqlite/execute! "/tmp/foo.db" ["delete from foo"]))
(prn (sqlite/execute! "/tmp/foo.db" ["insert into foo (the_text, the_int, the_real) values (?,?,?)" "foo" "1" "3.14"]))
(def results (sqlite/query!   "/tmp/foo.db" ["select * from foo"]))
(prn results)

(deftest results-test
  (is (= [{:the_int 1, :the_real 3.14, :the_text "foo"}] results)))

(let [{:keys [:fail :error]} (t/run-tests)]
  (System/exit (+ fail error)))
