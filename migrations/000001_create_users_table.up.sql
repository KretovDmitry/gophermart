create table if not exists users (
	id serial primary key,
	login varchar(255) not null constraint unique_login unique,
	password varchar(255) not null,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP 
);
