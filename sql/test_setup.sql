create table articles (
  `id` int not null auto_increment,
  primary key (`id`)
) engine=innodb default charset=utf8;

insert into articles (id) values (1), (2), (3), (4), (5);
