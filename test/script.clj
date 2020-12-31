#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]
            [clojure.java.io :as io]
            [clojure.test :as t :refer [deftest is]]))

(prn (pods/load-pod "./pod-babashka-sqlite3"))

(require '[pod.babashka.sqlite3 :as sqlite])

(.delete (io/file "/tmp/foo.db"))

(prn (sqlite/execute! "/tmp/foo.db" ["create table if not exists foo (the_text TEXT, the_int INTEGER, the_real REAL, the_blob BLOB)"]))
(prn (sqlite/execute! "/tmp/foo.db" ["delete from foo"]))

(def png (java.nio.file.Files/readAllBytes (.toPath (io/file "resources/babashka.png"))))

(prn (sqlite/execute! "/tmp/foo.db" ["insert into foo (the_text, the_int, the_real, the_blob) values (?,?,?,?)" "foo" 1 3.14 png]))

(def results (sqlite/query! "/tmp/foo.db" ["select * from foo"]))

(def results-min-png (update results 0 #(dissoc % "the_blob")))
(prn results-min-png)

(deftest results-test
  (is (= [{:the_int 1, :the_real 3.14, :the_text "foo"}] results-min-png)))

;; TODO:
(prn (count png) (count (get-in results [0 "the_blob"])))

(let [{:keys [:fail :error]} (t/run-tests)]
  (System/exit (+ fail error)))
