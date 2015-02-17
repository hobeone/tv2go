angular.module('shows.create', ['tv2go.models.shows','tv2go.indexerSearchService'
])
.config(function($stateProvider){
  $stateProvider
  .state('tv2go.shows.create', {
    url: '/create',
    views: {
      'detail@tv2go': {
        templateUrl: 'shows/create/shows-create.tmpl.html',
        controller: 'CreateShowsCtrl as createShowsCtrl'
      },
    },
  })
  .state('tv2go.shows.create.stepone', {
    url: '/stepone',
    templateUrl: 'shows/create/create-stepone.html',
  })
  .state('tv2go.shows.create.steptwo', {
    url: '/steptwo',
    templateUrl: 'shows/create/create-steptwo.html',
  })
  ;
})
.controller('CreateShowsCtrl', function($state, Show, ShowsModel, IndexerSearch) {
  var createShowsCtrl = this;

  function resetForm() {
    createShowsCtrl.showSearchReqest = new IndexerSearch();
    createShowsCtrl.showSearchReqest.indexer_name =  'tvdb';
    createShowsCtrl.showSearchResult = {};


    createShowsCtrl.newShow = new Show();
  }

  function cancelCreating() {
    returnToShows();
  }

  function returnToShows(){
    $state.go('tv2go.shows', {});
  }

  function searchShow(show) {
    console.log(show);
    createShowsCtrl.showSearchResult = IndexerSearch.query(show, function(){
      createShowsCtrl.newShow.indexer_name = show.indexer_name;
      $state.go('tv2go.shows.create.steptwo');
    },function(resp) {
      console.log(resp);
      window.alert(resp.statusText);
    });
  }

  function createShow(show) {
    ShowsModel.createShow(show);
    returnToShows();
  }

  createShowsCtrl.searchShow = searchShow;
  createShowsCtrl.cancelCreating = cancelCreating;
  createShowsCtrl.createShow = createShow;

  resetForm();
});
