package internal

import (
	"github.com/supabase-community/supabase-go"
)

var Supabase *supabase.Client

func InitSupabase(url, key string) (*supabase.Client, error) {
	return supabase.NewClient(url, key, nil)
}
