import ProblemEditPage from '../../[id]/edit/page';

export default function NewProblemEditPage() {
  return <ProblemEditPage params={Promise.resolve({ id: 'new' })} />;
}
