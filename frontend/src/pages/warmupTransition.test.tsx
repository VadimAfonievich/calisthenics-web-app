// @vitest-environment jsdom
import {act,cleanup,fireEvent,render,screen} from '@testing-library/react'
import {afterEach,describe,expect,it,vi} from 'vitest'
import {WarmupTransition} from './workout'

afterEach(cleanup)
describe('warmup to main transition',()=>{
  it('opens the persisted main session exactly once without a second countdown',async()=>{const open=vi.fn();render(<WarmupTransition sessionID="main-session" title="Корпус и ноги" onContinue={open}/>);await act(async()=>Promise.resolve());expect(open).toHaveBeenCalledTimes(1);expect(open).toHaveBeenCalledWith('main-session');expect(screen.queryByText(/Начинаем через/)).toBeNull()})
  it('guards the automatic/manual race',async()=>{let resolve!:()=>void;const open=vi.fn(()=>new Promise<void>(done=>{resolve=done}));render(<WarmupTransition sessionID="main-session" title="Main" onContinue={open}/>);await act(async()=>Promise.resolve());fireEvent.click(screen.getByRole('button'));expect(open).toHaveBeenCalledTimes(1);resolve()})
  it('shows a controlled error and retries the same session',async()=>{const open=vi.fn().mockRejectedValueOnce(new Error('route failed')).mockResolvedValueOnce(undefined);render(<WarmupTransition sessionID="same-session" title="Main" onContinue={open}/>);expect(await screen.findByText('Не удалось открыть основную тренировку.')).toBeTruthy();fireEvent.click(screen.getByRole('button',{name:'Повторить'}));expect(open).toHaveBeenCalledTimes(2);expect(open).toHaveBeenLastCalledWith('same-session')})
})
