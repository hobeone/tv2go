angular.module('shows.episodes',[
  'shows.episodes.create',
  'shows.episodes.edit',
  'tv2go.models.shows',
  'tv2go.models.episodes',
])
.config(function($stateProvider){
  $stateProvider
    .state('tv2go.shows.episodes', {
      url: "shows/:show",
      views: {
        'episodes@': {
          templateUrl: 'shows/episodes/episodes.tmpl.html',
          controller: 'EpisodesListCtrl as episodesListCtrl',
        }
      },
    });
})
.controller('EpisodesListCtrl', function ($stateParams, EpisodesModel, ShowsModel) {
  var EpisodesListCtrl = this;

  ShowsModel.setCurrentShow($stateParams.show);

  EpisodesModel.getEpisodes()
  .then(function(episodes){
    EpisodesListCtrl.episodes = episodes;
  });

  EpisodesListCtrl.getCurrentShow = ShowsModel.getCurrentShow;
  EpisodesListCtrl.getCurrentShowName = ShowsModel.getCurrentShowName;
  EpisodesListCtrl.deleteEpisode = EpisodesModel.deleteEpisode;
})
;
