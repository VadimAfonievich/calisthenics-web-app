// @vitest-environment jsdom
import {cleanup,render,screen} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {MemoryRouter} from 'react-router-dom'
import {afterEach,beforeEach,describe,expect,it} from 'vitest'
import {useSessionStore} from '../store/session'
import {ProfilePage} from './profile'

const student={id:'x',first_name:'Вадим',display_name:'Вадим',username:'vadim',level:4,xp:850,current_streak:8,timezone:'UTC',role:'user' as const,available_modes:['student' as const]}
const coach={...student,role:'coach' as const,available_modes:['student' as const,'coach' as const]}
const view=()=>render(<MemoryRouter><ProfilePage/></MemoryRouter>)
afterEach(cleanup)
describe('Profile application modes',()=>{
  beforeEach(()=>{localStorage.clear();useSessionStore.setState({status:'authenticated',user:student,appMode:'student'})})
  it('shows only student mode to a normal user',()=>{view();expect(screen.getByRole('button',{name:/Ученик/})).toBeTruthy();expect(screen.queryByRole('button',{name:/Тренер/})).toBeNull()})
  it('shows both modes to a coach',()=>{useSessionStore.setState({user:coach});view();expect(screen.getByRole('button',{name:/Ученик/})).toBeTruthy();expect(screen.getByRole('button',{name:/Тренер/})).toBeTruthy()})
  it('switches coach back to student',async()=>{useSessionStore.setState({user:coach,appMode:'coach'});view();await userEvent.click(screen.getByRole('button',{name:/Ученик/}));expect(useSessionStore.getState().appMode).toBe('student')})
})
