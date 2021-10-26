create table posts (
	id serial primary key,
	old_id integer,

	author integer,

	post date,
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