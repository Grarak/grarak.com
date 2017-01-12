import { ProjectFolderPage } from './app.po';

describe('project-folder App', function() {
  let page: ProjectFolderPage;

  beforeEach(() => {
    page = new ProjectFolderPage();
  });

  it('should display message saying app works', () => {
    page.navigateTo();
    expect(page.getParagraphText()).toEqual('app works!');
  });
});
