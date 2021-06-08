create database microservices;

\c microservices

create table messages (
  id serial primary key,
  message text not null,
  created timestamp not null
);
