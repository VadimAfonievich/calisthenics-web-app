// @vitest-environment jsdom
import {QueryClient,QueryClientProvider} from '@tanstack/react-query'
import {cleanup,fireEvent,render,screen} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {afterEach,beforeEach,describe,expect,it,vi} from 'vitest'
import {useSessionStore} from '../store/session'
import {TenantSettingsPage} from './tenantSettings'

const tenant={id:'10000000-0000-0000-0000-000000000001',slug:'old-school',name:'AFONICH',role:'coach' as const,description:'Сильнее каждый день',status:'active'}
const mocks=vi.hoisted(()=>({getCoachSpace:vi.fn(),updateCoachSpace:vi.fn(),updateCoachSpaceSlug:vi.fn(),updateCoachSpaceAvatar:vi.fn(),copyShareLink:vi.fn(),addExternalMedia:vi.fn()}))
vi.mock('../api/tenants',()=>({...mocks,tenantShareLink:(slug:string)=>`https://t.me/test_bot?startapp=${slug}`}))
vi.mock('../api/coach',()=>({addExternalMedia:mocks.addExternalMedia}))

beforeEach(()=>{mocks.getCoachSpace.mockResolvedValue({tenant});mocks.updateCoachSpace.mockResolvedValue({tenant});mocks.updateCoachSpaceSlug.mockImplementation(async(_token:string,slug:string)=>({tenant:{...tenant,slug}}));mocks.addExternalMedia.mockResolvedValue({media:{id:'20000000-0000-0000-0000-000000000001'}});mocks.updateCoachSpaceAvatar.mockResolvedValue({tenant:{...tenant,avatar_media_id:'20000000-0000-0000-0000-000000000001',avatar_url:'data:image/jpeg;base64,eA=='}});useSessionStore.setState({accessToken:'token',status:'authenticated',user:{id:'u',first_name:'Тренер',display_name:'Тренер',level:1,xp:0,current_streak:0,timezone:'UTC',role:'user',available_modes:['student','coach'],current_tenant:tenant}})})
afterEach(()=>{cleanup();vi.clearAllMocks()})

describe('настройки школы',()=>{
  it('uses clear Russian editable labels without technical Slug terminology',async()=>{render(<QueryClientProvider client={new QueryClient()}><TenantSettingsPage/></QueryClientProvider>);expect((await screen.findByLabelText('Название школы')).classList.contains('editable-control')).toBe(true);expect(screen.getByLabelText('Описание для учеников').getAttribute('placeholder')).toBeTruthy();expect(screen.getByText(/Этот текст будет виден ученикам/)).toBeTruthy();expect(screen.getByLabelText('Адрес школы')).toBeTruthy();expect(screen.queryByText(/^Slug$/)).toBeNull()})
  it('confirms slug change and updates the share link after success',async()=>{const user=userEvent.setup();render(<QueryClientProvider client={new QueryClient()}><TenantSettingsPage/></QueryClientProvider>);const address=await screen.findByLabelText('Адрес школы');await user.clear(address);await user.type(address,'new-school');await user.click(screen.getByRole('button',{name:'Изменить адрес'}));expect(screen.getByRole('alertdialog').textContent).toContain('После изменения старые ссылки');await user.click(screen.getByRole('alertdialog').querySelector('button:last-child')!);expect((await screen.findAllByText('https://t.me/test_bot?startapp=new-school')).length).toBe(2);expect(mocks.updateCoachSpaceSlug).toHaveBeenCalledWith('token','new-school')})
  it('uploads a mobile photo and reports completion',async()=>{const user=userEvent.setup();render(<QueryClientProvider client={new QueryClient()}><TenantSettingsPage/></QueryClientProvider>);const input=await screen.findByLabelText('Выбор фотографии школы');await user.upload(input,new File(['x'],'school.jpg',{type:'image/jpeg'}));expect(await screen.findByText('Аватар школы обновлён')).toBeTruthy();expect(mocks.addExternalMedia).toHaveBeenCalledWith('token',expect.objectContaining({type:'image',mime_type:'image/jpeg',original_filename:'school.jpg'}));expect(mocks.updateCoachSpaceAvatar).toHaveBeenCalledWith('token','20000000-0000-0000-0000-000000000001')})
  it('shows a clear error for an unsupported phone photo',async()=>{render(<QueryClientProvider client={new QueryClient()}><TenantSettingsPage/></QueryClientProvider>);fireEvent.change(await screen.findByLabelText('Выбор фотографии школы'),{target:{files:[new File(['x'],'school.heic',{type:'image/heic'})]}});expect((await screen.findByRole('alert')).textContent).toContain('JPEG, PNG и WebP');expect(mocks.addExternalMedia).not.toHaveBeenCalled()})
})
