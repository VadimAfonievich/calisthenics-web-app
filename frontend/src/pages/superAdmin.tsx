import {useState} from 'react'
import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'
import {copyShareLink, listTenants, tenantShareLink} from '../api/tenants'
import {searchUsers, setCoachRole, type RoleUser} from '../api/superAdmin'
import {useSessionStore} from '../store/session'

const safeSlug = /^[a-z0-9]+(?:-[a-z0-9]+)*$/

export function SuperAdminUsersPage() {
  const token = useSessionStore((state) => state.accessToken)!
  const me = useSessionStore((state) => state.user)
  const [tab, setTab] = useState<'users' | 'tenants'>('users')
  const [search, setSearch] = useState('')
  const [target, setTarget] = useState<RoleUser>()
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [created, setCreated] = useState<{name: string; slug: string}>()
  const client = useQueryClient()

  const users = useQuery({
    queryKey: ['super-admin-users', search],
    queryFn: () => searchUsers(token, search),
    enabled: me?.role === 'super_admin' && tab === 'users',
  })
  const tenants = useQuery({
    queryKey: ['super-admin-tenants'],
    queryFn: () => listTenants(token),
    enabled: me?.role === 'super_admin' && tab === 'tenants',
  })
  const role = useMutation({
    mutationFn: ({id, value, space}: {id: string; value: 'user' | 'coach'; space?: {name: string; slug: string}}) =>
      setCoachRole(token, id, value, space),
    onSuccess: (_, variables) => {
      void client.invalidateQueries({queryKey: ['super-admin-users']})
      void client.invalidateQueries({queryKey: ['super-admin-tenants']})
      if (variables.space) setCreated(variables.space)
      setTarget(undefined)
    },
  })

  if (me?.role !== 'super_admin') return <div className="notice error">Доступно только super_admin.</div>

  return <div className="stack admin-tenants-page">
    <div><p className="eyebrow">АДМИНИСТРИРОВАНИЕ</p><h2>{tab === 'users' ? 'Пользователи' : 'Пространства'}</h2></div>
    <div className="mode-switch">
      <button className={tab === 'users' ? 'active' : ''} onClick={() => setTab('users')}>Пользователи</button>
      <button className={tab === 'tenants' ? 'active' : ''} onClick={() => setTab('tenants')}>Пространства</button>
    </div>

    {tab === 'users' ? <>
      <input type="search" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Имя, @username или Telegram ID" />
      {users.data?.users.map((user) => <article className="card user-role-card" key={user.id}>
        <div><h3>{user.display_name}</h3>{user.username && <p>@{user.username}</p>}<small>Роль: {user.role === 'coach' ? 'Тренер' : user.role === 'user' ? 'Пользователь' : user.role}</small></div>
        {user.id !== me.id && (user.role === 'user' || user.role === 'coach') && <button disabled={role.isPending} onClick={() => {
          if (user.role === 'coach') role.mutate({id: user.id, value: 'user'})
          else {
            setTarget(user)
            setName(`${user.display_name} · школа`)
            setSlug(user.username?.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') ?? '')
          }
        }}>{user.role === 'coach' ? 'Снять роль тренера' : 'Сделать тренером'}</button>}
      </article>)}

      {target && <section className="card stack promotion-form">
        <h3>Создать пространство для {target.display_name}</h3>
        <label>Название пространства<input value={name} onChange={(event) => setName(event.target.value)} /></label>
        <label>Slug<input value={slug} onChange={(event) => setSlug(event.target.value.toLowerCase().replace(/\s+/g, '-'))} /></label>
        <button disabled={name.trim().length < 2 || !safeSlug.test(slug)} onClick={() => role.mutate({id: target.id, value: 'coach', space: {name: name.trim(), slug}})}>Назначить тренером</button>
        <button onClick={() => setTarget(undefined)}>Отмена</button>
      </section>}
      {created && <section className="notice share-card"><b>Тренер назначен.</b><p>Пространство: {created.name}</p><a href={tenantShareLink(created.slug)}>{tenantShareLink(created.slug)}</a><button onClick={() => copyShareLink(created.slug)}>Скопировать</button></section>}
    </> : <>
      {tenants.data?.tenants.map((tenant) => <article className="card tenant-admin-card" key={tenant.id}>
        <h3>{tenant.name}</h3><p className="tenant-slug">{tenant.slug}</p>
        <small>Владелец: {tenant.owner_name} · Статус: {tenant.status} · Учеников: {tenant.students ?? 0}</small>
        {tenant.created_at && <small>Создано: {new Date(tenant.created_at).toLocaleString('ru-RU')}</small>}
        <a href={tenantShareLink(tenant.slug)}>{tenantShareLink(tenant.slug)}</a>
      </article>)}
    </>}
  </div>
}
