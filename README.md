# go-crud-api

:warning: Work in progress :warning: attempt to port [php-crud-api](https://github.com/mevdschee/php-crud-api) to golang.

To-do :
- [ ] Tests : 
  - [ ] more unit tests
  - [X] implement php-crud-api tests
- [X] Other drivers (only sqlite now)
- [X] Cache mecanism
- [X] Finishing controllers
- [ ] Custom controller (compile extra go code at launch like https://github.com/benhoyt/prig ?)
- [X] Finishing middlewares
- [ ] Add a github workflow <- next
  - [X] Init
  - [ ] Add pgql, mysql and sqlserver testing
  - [ ] Find why somes linters are not working
  - [ ] Release pipeline
- [ ] Add an alter table function for sqlite (create new table, copy data, drop old table)
- [ ] Review whole package structure
- [ ] Logger options
- [X] https
- [ ] Write a README <- next
- [ ] Comment code
- [ ] :tada: