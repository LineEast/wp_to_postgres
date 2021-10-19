create table posts (
	id serial primary key,
	old_id integer,

	author integer,

	post date,
	content text,
	title varchar(200),
	img varchar(200),
	
	tags_id jsonb default '[]'
)

create table tags (
	id serial primary key,
	name varchar(200),
	slug varchar(200),
	count integer
)