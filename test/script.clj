#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]
            [clojure.java.io :as io]
            [clojure.test :as t :refer [deftest is testing]]))

(prn (pods/load-pod "./pod-babashka-go-sqlite3"))

(require '[pod.babashka.go-sqlite3 :as sqlite])

(.delete (io/file "/tmp/foo.db"))

(prn (sqlite/execute! "/tmp/foo.db" ["create table if not exists foo (the_text TEXT, the_int INTEGER, the_real REAL, the_blob BLOB)"]))
(prn (sqlite/execute! "/tmp/foo.db" ["delete from foo"]))

(def png (java.nio.file.Files/readAllBytes (.toPath (io/file "resources/babashka.png"))))

(prn (sqlite/execute! "/tmp/foo.db" ["insert into foo (the_text, the_int, the_real, the_blob) values (?,?,?,?)" "foo" 1 3.14 png]))
(prn (sqlite/execute! "/tmp/foo.db" ["insert into foo (the_text, the_int, the_real) values (?,?,?)" "foo" 2 1.5]))

(testing "multiple results"
  (prn (sqlite/execute! "/tmp/foo.db"
                        ["insert into foo (the_text, the_int, the_real) values (?,?,?), (?,?,?)"
                         "bar" 3 1.5
                         "baz" 4 1.5])))

(def results (sqlite/query "/tmp/foo.db" ["select * from foo order by the_int asc"]))

(def results-min-png (mapv #(dissoc % :the_blob) results))

(deftest results-test
  (is (= [{:the_int 1, :the_real 3.14, :the_text "foo"}
          {:the_int 2, :the_real 1.5, :the_text "foo"}
          {:the_int 3, :the_real 1.5, :the_text "bar"}
          {:the_int 4, :the_real 1.5, :the_text "baz"}]
         results-min-png)))

(deftest bytes-roundtrip
  (is (= (count png) (count (get-in results [0 :the_blob])))))

(deftest error-test
  (is (thrown-with-msg?
       Exception #"no such column: non_existing"
       (sqlite/query "/tmp/foo.db" ["select non_existing from foo"])))
  (is (thrown-with-msg?
       Exception #"expected query to be a vector"
       (sqlite/query "/tmp/foo.db" "select * from foo"))))

(let [{:keys [:fail :error]} (t/run-tests)]
  (System/exit (+ fail error)))
