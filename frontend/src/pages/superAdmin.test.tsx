// @vitest-environment jsdom
import {QueryClient,QueryClientProvider} from '@tanstack/react-query'
import {cleanup,render,screen} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {afterEach,beforeEach,expect,it,vi} from 'vitest'
import {useSessionStore} from '../store/session'
import {SuperAdminUsersPage} from './superAdmin'
const mocks=vi.hoisted(()=>({searchUsers:vi.fn(),setCoachRole:vi.fn()}));vi.mock('../api/superAdmin',()=>mocks)
beforeEach(()=>{useSessionStore.setState({accessToken:'token',user:{id:'admin',display_name:'Root',first_name:'Root',level:1,xp:0,current_streak:0,timezone:'UTC',role:'super_admin',available_modes:['student','coach']},status:'authenticated'});mocks.searchUsers.mockResolvedValue({users:[{id:'10000000-0000-0000-0000-000000000001',telegram_id:1,username:'ivan',display_name:'Иван',role:'user'}]});mocks.setCoachRole.mockResolvedValue({role:'coach'})});afterEach(cleanup)
it('searches and promotes a user to coach',async()=>{render(<QueryClientProvider client={new QueryClient()}><SuperAdminUsersPage/></QueryClientProvider>);expect(await screen.findByText('Иван')).toBeTruthy();await userEvent.click(screen.getByText('Сделать тренером'));expect(mocks.setCoachRole).toHaveBeenCalledWith('token','10000000-0000-0000-0000-000000000001','coach')})
