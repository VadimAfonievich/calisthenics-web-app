// @vitest-environment jsdom
import {render,screen} from '@testing-library/react'
import {MemoryRouter,Route,Routes} from 'react-router-dom'
import {beforeEach,describe,expect,it} from 'vitest'
import {useSessionStore} from '../store/session'
import {CoachGuard} from './coachGuard'

const base={id:'x',first_name:'A',display_name:'A',level:1,xp:0,current_streak:0,timezone:'UTC',role:'user' as const,available_modes:['student' as const]}
const renderGuard=()=>render(<MemoryRouter initialEntries={['/coach']}><Routes><Route path="/coach" element={<CoachGuard><div>coach-area</div></CoachGuard>}/><Route path="/profile" element={<div>profile-area</div>}/></Routes></MemoryRouter>)
describe('CoachGuard',()=>{
  beforeEach(()=>useSessionStore.setState({status:'authenticated',user:base,appMode:'student'}))
  it('denies a normal user opening /coach directly',()=>{renderGuard();expect(screen.getByText('profile-area')).toBeTruthy()})
  it('allows an authorized coach',()=>{useSessionStore.setState({user:{...base,role:'coach',available_modes:['student','coach']}});renderGuard();expect(screen.getByText('coach-area')).toBeTruthy()})
})
