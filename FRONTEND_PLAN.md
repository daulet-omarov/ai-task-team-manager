# Frontend Plan — AI Task Team Manager

## Stack

| Библиотека | Назначение |
|---|---|
| React + TypeScript | UI framework |
| React Router v6 | Роутинг |
| TanStack Query | Server state, кэш, loading/error |
| Zustand | Client state (токен, текущий пользователь) |
| Axios | HTTP клиент |
| shadcn/ui + Tailwind CSS | UI компоненты |
| React Hook Form + Zod | Формы и валидация |

---

## Структура проекта

```
src/
├── api/                  # функции запросов к API
│   ├── auth.ts
│   ├── board.ts
│   ├── employee.ts
│   ├── invite.ts
│   └── task.ts
├── components/           # переиспользуемые компоненты
│   ├── ui/               # shadcn компоненты (кнопки, модалки, инпуты)
│   ├── BoardCard.tsx
│   ├── TaskCard.tsx
│   ├── TaskModal.tsx
│   ├── InviteModal.tsx
│   └── KanbanColumn.tsx
├── pages/
│   ├── LoginPage.tsx
│   ├── RegisterPage.tsx
│   ├── VerifyEmailPage.tsx
│   ├── ForgotPasswordPage.tsx
│   ├── ResetPasswordPage.tsx
│   ├── OnboardingPage.tsx
│   ├── DashboardPage.tsx
│   ├── BoardPage.tsx
│   ├── InvitesPage.tsx
│   └── ProfilePage.tsx
├── store/
│   └── authStore.ts      # Zustand: token, user_id, full_name
├── hooks/                # кастомные хуки поверх TanStack Query
│   ├── useAuth.ts
│   ├── useBoards.ts
│   ├── useTasks.ts
│   └── useInvites.ts
├── lib/
│   ├── axios.ts          # инстанс axios с interceptors
│   └── utils.ts
└── router.tsx            # описание всех маршрутов
```

---

## Страницы и маршруты

```
/login                    — вход
/register                 — регистрация
/verify-email?token=...   — верификация email (редирект из письма)
/forgot-password          — запрос сброса пароля
/reset-password?token=... — форма нового пароля

/onboarding               — заполнить профиль (если isFirstLogin = true)
/dashboard                — главная (список досок)
/boards/:id               — доска с задачами (Kanban)
/boards/:id/settings      — настройки доски (участники, инвайты)
/invites                  — входящие приглашения
/profile                  — профиль / редактирование
```

### Защищённые маршруты

```
Нет токена → редирект на /login
Есть токен → пропускаем дальше
```

---

## Auth Flow

```
POST /auth/login
  ↓
Сохранить в localStorage + Zustand:
  { token, user_id, full_name, email }
  ↓
GET /employees/exists
  ↓
exists = false → /onboarding
exists = true  → /dashboard
```

```
Axios interceptor (src/lib/axios.ts):
  - Каждый запрос: добавляет заголовок Authorization: Bearer <token>
  - Ответ 401: очищает токен → редирект /login
```

---

## API слой

### `api/auth.ts`

```ts
register(email: string, password: string): Promise<void>
login(email: string, password: string): Promise<LoginResponse>
forgotPassword(email: string): Promise<void>
resetPassword(token: string, newPassword: string): Promise<void>
deleteAccount(): Promise<void>

type LoginResponse = {
  token: string
  user_id: string
  full_name: string | null
  email: string
}
```

### `api/employee.ts`

```ts
createEmployee(data: CreateEmployeeRequest): Promise<void>
getMe(): Promise<EmployeeResponse>
updateEmployee(data: UpdateEmployeeRequest): Promise<void>
deleteEmployee(): Promise<void>
exists(): Promise<{ exists: boolean }>
getAll(): Promise<EmployeeResponse[]>

type EmployeeResponse = {
  id: number
  user_id: number
  full_name: string
  photo: string
  email: string
  phone_number: string
  birthday: string
  team: { id: number; name: string; code: string }
  gender: { id: number; name: string; code: string }
}
```

### `api/board.ts`

```ts
getDashboard(): Promise<DashboardResponse>
createBoard(data: CreateBoardRequest): Promise<BoardResponse>
getBoard(id: number): Promise<BoardResponse>

type DashboardResponse = {
  boards: BoardResponse[]
  isFirstLogin: boolean
}

type BoardResponse = {
  id: number
  name: string
  description: string
  memberCount: number
  isOwner: boolean
  ownerId: number
}
```

### `api/task.ts`

```ts
createTask(boardId: number, data: CreateTaskRequest): Promise<TaskResponse>
getTask(taskId: number): Promise<TaskResponse>
updateTask(taskId: number, data: UpdateTaskRequest): Promise<TaskResponse>
getBoardTasks(boardId: number): Promise<TaskResponse[]>  // ← нужен новый эндпойнт

type TaskResponse = {
  id: number
  board_id: number
  title: string
  description: string
  status_id: number
  priority_id: number
  assignee_id: number
  reporter_id: number
  time_spent: number
  created_at: string
  updated_at: string
}
```

### `api/invite.ts`

```ts
sendInvite(boardId: number, userId: number): Promise<void>
getInvites(): Promise<InviteResponse[]>
acceptInvite(inviteId: number): Promise<void>
rejectInvite(inviteId: number): Promise<void>

type InviteResponse = {
  id: number
  board_id: number
  board_name: string
  inviter_id: number
  invitee_id: number
  status: 'pending' | 'accepted' | 'rejected'
  created_at: string
}
```

---

## Страницы

### `/login` и `/register`

- Форма (React Hook Form + Zod)
- После логина — auth flow (см. выше)
- После регистрации — сообщение "проверьте почту"

### `/verify-email?token=...`

- При открытии сразу делает `GET /auth/verify-email?token=...`
- Успех → редирект `/login` с сообщением
- Ошибка → "ссылка недействительна"

### `/onboarding`

```
Заполнить профиль (показывается только если isFirstLogin = true)

Поля: full_name, email, birthday, phone, team, gender
POST /employees

После создания → /dashboard
```

### `/dashboard`

```
GET /dashboard
  ↓
isFirstLogin = true  → редирект /onboarding
isFirstLogin = false → рендер

┌──────────────────────────────────────┐
│  My Boards                  [+ New]  │
│                                      │
│  ┌──────────┐  ┌──────────┐          │
│  │ Board A  │  │ Board B  │  ...     │
│  │ 5 members│  │ 2 members│          │
│  │ 👑 Owner │  │ Member   │          │
│  └──────────┘  └──────────┘          │
│                                      │
│  Invites (3 pending)  →              │
└──────────────────────────────────────┘
```

- Клик по карточке → `/boards/:id`
- Кнопка `+ New` → модалка создания доски
- Бейдж с кол-вом инвайтов → `/invites`

### `/boards/:id` — Kanban

```
GET /boards/:id
GET /boards/:id/tasks   ← нужно добавить в backend

┌──────────────────────────────────────────────┐
│  Board Name                  [Invite] [⚙]    │
│                                              │
│  TODO          IN PROGRESS       DONE        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Task 1   │  │ Task 3   │  │ Task 5   │   │
│  │ #high    │  │ @daulet  │  │          │   │
│  ├──────────┤  └──────────┘  └──────────┘   │
│  │ Task 2   │                               │
│  └──────────┘                               │
│  [+ Add task]                               │
└──────────────────────────────────────────────┘
```

- Клик по задаче → `TaskModal`
- `[+ Add task]` → форма создания задачи
- `[Invite]` → `InviteModal` (поиск по email из `GET /employees`)

### `TaskModal`

```
GET /tasks/:id

Поля для редактирования:
  title, description, status (select), priority (select),
  assignee (select из участников доски), time_spent

PATCH /tasks/:id  при изменении любого поля (debounce или кнопка Save)
```

### `/invites`

```
GET /invites

┌──────────────────────────────────────┐
│  Invitations                         │
│                                      │
│  "Board Alpha"  from user #5         │
│  [Accept ✓]  [Reject ✗]             │
│                                      │
│  "Board Beta"   from user #2         │
│  [Accept ✓]  [Reject ✗]             │
└──────────────────────────────────────┘

Accept → POST /invites/:id/accept → убрать из списка → обновить dashboard
Reject → POST /invites/:id/reject → убрать из списка
```

### `/profile`

```
GET /employees/me

Форма редактирования профиля:
  full_name, email, phone, birthday, team, gender, photo

PUT /employees  при сохранении
```

---

## Zustand Store

```ts
// store/authStore.ts
type AuthState = {
  token: string | null
  userId: string | null
  fullName: string | null
  email: string | null
  setAuth: (data: LoginResponse) => void
  clearAuth: () => void
}
```

Данные сохраняются в `localStorage` через `zustand/middleware/persist`.

---

## Что нужно добавить в backend

| Эндпойнт | Зачем нужен |
|---|---|
| `GET /boards/:id/tasks` | Загрузить все задачи доски для Kanban |

> `GET /employees` уже есть — используется для поиска участников при инвайте.

---

## Порядок реализации

```
Этап 1 — Auth
  ✦ /login, /register
  ✦ Axios interceptor + Zustand
  ✦ /verify-email, /forgot-password, /reset-password

Этап 2 — Onboarding + Dashboard
  ✦ /onboarding (форма профиля)
  ✦ /dashboard (список досок, создание)

Этап 3 — Invites
  ✦ /invites (список, принять, отклонить)
  ✦ Бейдж на dashboard

Этап 4 — Board + Tasks
  ✦ /boards/:id (Kanban)
  ✦ TaskModal (просмотр + редактирование)
  ✦ Создание задачи

Этап 5 — Invite внутри доски
  ✦ InviteModal (поиск участников по email)
  ✦ POST /boards/:id/invite

Этап 6 — Profile
  ✦ /profile (просмотр + редактирование)
```
