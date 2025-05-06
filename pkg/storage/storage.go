// Пакет для работы с БД приложения GoNews.
package storage

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

// База данных.
type DB struct {
	pool *pgxpool.Pool
}

// Публикация, получаемая из RSS.
type Post struct {
	ID      int    // номер записи
	Title   string // заголовок публикации
	Content string // содержание публикации
	PubTime int64  // время публикации
	Link    string // ссылка на источник
}

func New() (*DB, error) {
	constr := "postgresql://postgres:admin@192.168.1.165/newsdb"
	pool, err := pgxpool.Connect(context.Background(), constr)
	if err != nil {
		return nil, err
	}
	db := DB{
		pool: pool,
	}
	return &db, nil
}

// StoreNews Добавляет новости в базу
func (db *DB) StoreNews(news []Post) error {
	for _, post := range news {
		_, err := db.pool.Exec(context.Background(), `
		INSERT INTO news(title, content, pub_time, link)
		VALUES ($1, $2, $3, $4)`,
			post.Title,
			post.Content,
			post.PubTime,
			post.Link,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// News возвращает последние новости из БД.
func (db *DB) News(n int) ([]Post, error) {
	if n == 0 {
		n = 10
	}
	rows, err := db.pool.Query(context.Background(), `
	SELECT id, title, content, pub_time, link FROM news
	ORDER BY pub_time DESC
	LIMIT $1
	`,
		n,
	)
	if err != nil {
		return nil, err
	}
	var news []Post
	for rows.Next() {
		var p Post
		err = rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.PubTime,
			&p.Link,
		)
		if err != nil {
			return nil, err
		}
		news = append(news, p)
	}
	return news, rows.Err()
}

// GetAllNews получения списка всех новостей
func (db *DB) GetAllNews() ([]Post, error) {
	rows, err := db.pool.Query(context.Background(), `
		SELECT id, title, content, pub_time, link FROM news ORDER BY pub_time DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.PubTime, &post.Link)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (db *DB) AllNewsPaginated(offset, limit int) ([]Post, int, error) {
	rows, err := db.pool.Query(context.Background(), `
		SELECT id, title, content, pub_time, link 
		FROM news 
		ORDER BY pub_time DESC
		OFFSET $1 LIMIT $2
	`, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.PubTime, &post.Link)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, post)
	}

	// Получаем общее количество записей
	var total int
	err = db.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM news`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (db *DB) SearchNews(query string, offset, limit int) ([]Post, int, error) {
	query = "%" + query + "%"
	rows, err := db.pool.Query(context.Background(), `
		SELECT id, title, content, pub_time, link 
		FROM news 
		WHERE title ILIKE $1 
		ORDER BY pub_time DESC
		OFFSET $2 LIMIT $3
	`, query, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.PubTime, &post.Link)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, post)
	}

	// Получаем общее количество записей по поиску
	var total int
	err = db.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM news WHERE title ILIKE $1
	`, query).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}
