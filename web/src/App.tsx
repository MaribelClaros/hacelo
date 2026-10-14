import { Routes, Route } from 'react-router-dom';
import TaskList from './pages/TaskList';
import NewTask from './pages/NewTask';
import Reschedule from './pages/Reschedule';
import Settings from './pages/Settings';

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<TaskList />} />
      <Route path="/tasks/new" element={<NewTask />} />
      <Route path="/tasks/:id/reschedule" element={<Reschedule />} />
      <Route path="/settings" element={<Settings />} />
    </Routes>
  );
}
