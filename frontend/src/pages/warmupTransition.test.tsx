// @vitest-environment jsdom
import {act,cleanup,fireEvent,render,screen} from '@testing-library/react'
import {afterEach,beforeEach,describe,expect,it,vi} from 'vitest'
import {WarmupTransition} from './workout'

beforeEach(()=>{vi.useFakeTimers();vi.setSystemTime(new Date('2026-08-21T00:00:00Z'))})
afterEach(()=>{cleanup();vi.useRealTimers()})

describe('warmup to main transition',()=>{
  it('automatically opens the persisted main session exactly once without rendering a stuck zero',async()=>{
    const open=vi.fn();render(<WarmupTransition sessionID="main-session" title="Корпус и ноги" onContinue={open}/>);
    expect(screen.getByText('Начинаем через 5')).toBeTruthy();
    await act(async()=>vi.advanceTimersByTime(5000));
    expect(open).toHaveBeenCalledTimes(1);expect(open).toHaveBeenCalledWith('main-session');expect(screen.queryByText('Начинаем через 0')).toBeNull();
    await act(async()=>vi.advanceTimersByTime(2000));expect(open).toHaveBeenCalledTimes(1);
  });
  it('starts immediately on tap and the elapsed countdown cannot navigate twice',async()=>{
    const open=vi.fn();render(<WarmupTransition sessionID="existing-session" title="Main" onContinue={open}/>);
    await act(async()=>vi.advanceTimersByTime(2000));expect(screen.getByText('Начинаем через 3')).toBeTruthy();
    fireEvent.click(screen.getByRole('button',{name:'Начать сейчас'}));expect(open).toHaveBeenCalledTimes(1);
    await act(async()=>vi.advanceTimersByTime(5000));expect(open).toHaveBeenCalledTimes(1);
  });
  it('guards an auto/manual race',async()=>{
    let resolve!:()=>void;const open=vi.fn(()=>new Promise<void>(done=>{resolve=done}));render(<WarmupTransition sessionID="main-session" title="Main" onContinue={open}/>);
    await act(async()=>vi.advanceTimersByTime(5000));fireEvent.click(screen.getByRole('button'));expect(open).toHaveBeenCalledTimes(1);resolve();
  });
  it('shows a controlled error and retries the same session',async()=>{
    const open=vi.fn().mockRejectedValueOnce(new Error('route failed')).mockResolvedValueOnce(undefined);render(<WarmupTransition sessionID="same-session" title="Main" onContinue={open}/>);
    fireEvent.click(screen.getByRole('button',{name:'Начать сейчас'}));await act(async()=>Promise.resolve());expect(screen.getByText('Не удалось открыть основную тренировку.')).toBeTruthy();
    fireEvent.click(screen.getByRole('button',{name:'Повторить'}));expect(open).toHaveBeenCalledTimes(2);expect(open).toHaveBeenLastCalledWith('same-session');
  });
})
