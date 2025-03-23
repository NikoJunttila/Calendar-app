go install github.com/a-h/templ/cmd/templ@latest <br>
go install github.com/pressly/goose/v3/cmd/goose@latest <br>

tailwind watch works kinda scuffed. need to run npx tailwindcss -i app/assets/app.css -o ./public/assets/styles.css
after new styles added
<br>
userID := kit.Auth().(auth.Auth).UserID
<br>
TODO<br>
Calendar view allow delete and edit of entry<br>
Fix juhannus holiday<br>
FIX duplicate email signup error <br>

run command to start with docker <br>
sudo docker-compose up -d <br>
export COMPOSE_BAKE=true in ~/.bashrc for faster build stage
<br>
For Cloudflared tunnel hosting.
create new docker network where we connect all applications.
sudo docker network create app-network <br>
sudo docker run -d --name cloudflared --network app-network cloudflare/cloudflared:latest tunnel --no-autoupdate run --token yourtokenhere <br>


