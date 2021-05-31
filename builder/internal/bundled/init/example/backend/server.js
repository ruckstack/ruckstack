'use strict';

var cors = require('cors')
const express = require('express');

// Constants
const PORT = 8080;
const HOST = '0.0.0.0';

var process = require('process')
process.on('SIGINT', () => {
    console.info("Interrupted")
    process.exit(0)
})

let dbHost = process.env.POSTGRES_HOST;
if (!dbHost) {
    dbHost = 'postgresql';
}

const Pool = require('pg').Pool
const pool = new Pool({
    user: 'postgres',
    host: dbHost,
    database: 'postgres',
    password: process.env.POSTGRES_PASSWORD,
    port: 5432,
})

pool.query('create table if not exists notes (id serial primary key, title varchar(255) not null, content text)', (error, results) => {
    if (error) {
        throw error;
    }
    console.info("Notes table created");
})

// App
const app = express();
app.use(express.json())
app.use(cors())

app.get('/api/notes', (req, res) => {
    console.log("Requested GET /api/notes");

    pool.query('SELECT * FROM notes ORDER BY id DESC', (error, results) => {
        if (error) {
            throw error
        }
        res.status(200)
            .json(results.rows)
    })
});

app.post('/api/notes', (req, res) => {
    console.log("Requested POST /api/notes");

    pool.query({
        name: 'notes-insert',
        text: 'insert into notes (title, content) values ($1, $2)',
        values: [
            req.body.title,
            req.body.content
        ]
    }, (error, results) => {
        if (error) {
            throw error
        }
        res.status(200)
            .json(results.rows)
    })
});

app.listen(PORT, HOST);
console.log(`Running on http://${HOST}:${PORT}`);