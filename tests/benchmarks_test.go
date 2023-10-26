package tests_test

import "testing"

func BenchmarkVKBot(b *testing.B) {
	bot := NewVKBot()

	for i := 0; i < b.N; i++ {
		if user := bot.GetUser(User{
			UserID: 1,
		}); user.Error != nil {
			b.Fatal(user.Error)
		}
	}
}
