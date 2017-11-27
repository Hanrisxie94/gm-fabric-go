// Copyright 2017 Decipher Technology Studios LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// basic seed data
func seedData() {
	movies = append(movies, movie{
		Title:    "A New Hope",
		Director: "George Lucas",
		Release:  "May 25, 1977",
	})

	movies = append(movies, movie{
		Title:    "Empire Strikes Back",
		Director: "George Lucas",
		Release:  "May 21, 1980",
	})

	movies = append(movies, movie{
		Title:    "Return of The Jedi",
		Director: "George Lucas",
		Release:  "May 25, 1983",
	})

	movies = append(movies, movie{
		Title:    "The Phantom Menace",
		Director: "George Lucas",
		Release:  "May 19, 1999",
	})

	movies = append(movies, movie{
		Title:    "Attack of The Clones",
		Director: "George Lucas",
		Release:  "May 16, 2002",
	})

	movies = append(movies, movie{
		Title:    "Revenge of The Sith",
		Director: "George Lucas",
		Release:  "May 19, 2005",
	})

	movies = append(movies, movie{
		Title:    "The Force Awakens",
		Director: "J.J. Abrams",
		Release:  "December 18, 2015",
	})
}
