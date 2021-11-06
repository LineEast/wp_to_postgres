create table posts (
	id serial primary key,
	old_id integer,

	author integer,

	post timestamp,
	content text,
	content_short text,
	title varchar(200),
	img varchar(255),
	
	tags_id int[] default '{}' 
)

create table tags (
	id serial primary key,
	name varchar(200),
	slug varchar(200),
	count integer,
	taxonomy varchar(200)
)




create table posts (
	id bigserial primary key,
	old_id bigint,

	author_id bigint,

	title varchar(256),
	image varchar(256),
	date timestamp,
	description text,
	content text
);

create table tags (
	id bigserial primary key,
	name varchar(256),
	alias varchar(256),
	type varchar(256) check (taxonomy in ('post_tag', 'category'))
);

create table posts_tags (
    post_id bigint,
    tag_id bigint,

    primary key (post_id, tag_id),

    foreign key (post_id) references posts (id),
    foreign key (tag_id) references tags (id)
);