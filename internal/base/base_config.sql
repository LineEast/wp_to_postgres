drop table posts_tags;
drop table tags;
drop table posts;

create table posts (
	id bigserial primary key,
	old_id bigint,

	author_id bigint,

	title varchar(256),
	image varchar(256),
	date timestamp,
	description text,
	content text,

	views bigint
);

create table tags (
	id bigserial primary key,
	name varchar(256),
	alias varchar(256),
	type varchar(256)
);

create table posts_tags (
    post_id bigint,
    tag_id bigint,

    primary key (post_id, tag_id),

    foreign key (post_id) references posts (id),
    foreign key (tag_id) references tags (id)
);